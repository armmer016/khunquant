package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/cryptoquantumwave/khunquant/pkg/config"
	"github.com/cryptoquantumwave/khunquant/pkg/dca"
	"github.com/cryptoquantumwave/khunquant/pkg/providers/broker"
)

// ExecuteDCAOrderTool executes one DCA purchase for a plan.
// Unlike create_order, this tool is purpose-built for autonomous DCA execution:
// it bypasses the user-confirm gate and atomically records the trade in the DCA store.
type ExecuteDCAOrderTool struct {
	cfg   *config.Config
	store *dca.Store
}

func NewExecuteDCAOrderTool(cfg *config.Config, store *dca.Store) *ExecuteDCAOrderTool {
	return &ExecuteDCAOrderTool{cfg: cfg, store: store}
}

func (t *ExecuteDCAOrderTool) Name() string { return NameExecuteDCAOrder }

func (t *ExecuteDCAOrderTool) Description() string {
	return "Execute one DCA (Dollar Cost Averaging) purchase for a plan. Fetches the current market price, places a market buy order for the plan's configured quote amount, and records the execution. This tool is pre-authorized for DCA automation — no additional confirmation is needed when called from a scheduled DCA task."
}

func (t *ExecuteDCAOrderTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"plan_id": map[string]any{"type": "integer", "description": "ID of the DCA plan to execute."},
		},
		"required": []string{"plan_id"},
	}
}

func (t *ExecuteDCAOrderTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	planIDf, _ := args["plan_id"].(float64)
	planID := int64(planIDf)
	if planID <= 0 {
		return ErrorResult("plan_id is required")
	}

	plan, err := t.store.GetPlan(ctx, planID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("plan not found: %v", err))
	}
	if !plan.Enabled {
		return ErrorResult(fmt.Sprintf("DCA plan %d (%s) is disabled — enable it first with update_dca_plan", planID, plan.Name))
	}

	// Check end date.
	if plan.EndDate != nil && time.Now().After(*plan.EndDate) {
		return ErrorResult(fmt.Sprintf("DCA plan %d (%s) has expired (end date: %s)", planID, plan.Name, plan.EndDate.Format("2006-01-02")))
	}

	// Safety gates: permission + rate limit.
	if err := broker.CheckPermission(t.cfg, plan.Provider, plan.Account, config.ScopeTrade); err != nil {
		return ErrorResult(fmt.Sprintf("permission check failed: %v", err))
	}
	if !broker.DefaultLimiter.Allow(plan.Provider) {
		return ErrorResult(fmt.Sprintf("rate limit exceeded for provider %q — try again in a minute", plan.Provider))
	}

	p, err := broker.CreateProviderForAccount(plan.Provider, plan.Account, t.cfg)
	if err != nil {
		return t.recordFailure(ctx, plan, 0, fmt.Sprintf("failed to create provider: %v", err))
	}

	// Fetch current price to compute base amount.
	md, ok := p.(broker.MarketDataProvider)
	if !ok {
		return t.recordFailure(ctx, plan, 0, fmt.Sprintf("provider %q does not support market data", plan.Provider))
	}
	ticker, err := md.FetchTicker(ctx, plan.Symbol)
	if err != nil {
		return t.recordFailure(ctx, plan, 0, fmt.Sprintf("failed to fetch ticker for %s: %v", plan.Symbol, err))
	}
	if ticker.Last == nil {
		return t.recordFailure(ctx, plan, 0, fmt.Sprintf("ticker for %s has no last price", plan.Symbol))
	}
	currentPrice := *ticker.Last
	if currentPrice <= 0 {
		return t.recordFailure(ctx, plan, 0, fmt.Sprintf("invalid price %.8g for %s", currentPrice, plan.Symbol))
	}

	// Convert quote amount to base amount for market buy.
	baseAmount := plan.AmountPerOrder / currentPrice

	tp, ok := p.(broker.TradingProvider)
	if !ok {
		return t.recordFailure(ctx, plan, currentPrice, fmt.Sprintf("provider %q does not support order execution", plan.Provider))
	}

	order, err := tp.CreateOrder(ctx, plan.Symbol, "market", "buy", baseAmount, nil, nil)
	if err != nil {
		return t.recordFailure(ctx, plan, currentPrice, fmt.Sprintf("order placement failed: %v", err))
	}

	// Extract filled details from order response.
	orderID := ""
	if order.Id != nil {
		orderID = *order.Id
	}
	filledPrice := currentPrice
	if order.Average != nil && *order.Average > 0 {
		filledPrice = *order.Average
	} else if order.Price != nil && *order.Price > 0 {
		filledPrice = *order.Price
	}
	filledQty := baseAmount
	if order.Filled != nil && *order.Filled > 0 {
		filledQty = *order.Filled
	}
	actualQuote := filledQty * filledPrice
	feeQuote := 0.0
	if order.Fee.Cost != nil {
		feeQuote = *order.Fee.Cost
	}

	now := time.Now().UTC()
	exec := &dca.Execution{
		PlanID:         planID,
		ExecutedAt:     now,
		Symbol:         plan.Symbol,
		Provider:       plan.Provider,
		Account:        plan.Account,
		OrderID:        orderID,
		AmountQuote:    actualQuote,
		FilledPrice:    filledPrice,
		FilledQuantity: filledQty,
		FeeQuote:       feeQuote,
		Status:         "completed",
		CreatedAt:      now,
	}
	_, _ = t.store.SaveExecution(ctx, exec)
	_ = t.store.UpdatePlanStats(ctx, planID, actualQuote, filledQty)

	out := fmt.Sprintf("DCA order executed for plan %d (%s):\n", planID, plan.Name)
	out += fmt.Sprintf("  Symbol:    %s\n", plan.Symbol)
	out += fmt.Sprintf("  Order ID:  %s\n", orderID)
	out += fmt.Sprintf("  Price:     %.8g\n", filledPrice)
	out += fmt.Sprintf("  Qty:       %.8g %s\n", filledQty, plan.Symbol[:len(plan.Symbol)-len("/"+split(plan.Symbol))])
	out += fmt.Sprintf("  Spent:     %.4f\n", actualQuote)
	if feeQuote > 0 {
		out += fmt.Sprintf("  Fee:       %.6f\n", feeQuote)
	}
	return UserResult(out)
}

// recordFailure saves a failed execution record and returns an error result.
func (t *ExecuteDCAOrderTool) recordFailure(ctx context.Context, plan *dca.Plan, price float64, msg string) *ToolResult {
	now := time.Now().UTC()
	exec := &dca.Execution{
		PlanID:     plan.ID,
		ExecutedAt: now,
		Symbol:     plan.Symbol,
		Provider:   plan.Provider,
		Account:    plan.Account,
		AmountQuote: plan.AmountPerOrder,
		FilledPrice: price,
		Status:     "failed",
		ErrorMsg:   msg,
		CreatedAt:  now,
	}
	_, _ = t.store.SaveExecution(ctx, exec)
	return ErrorResult(fmt.Sprintf("DCA execution failed for plan %d (%s): %s", plan.ID, plan.Name, msg))
}

// split returns the quote currency part of a symbol like "BTC/USDT" → "USDT".
func split(symbol string) string {
	for i := len(symbol) - 1; i >= 0; i-- {
		if symbol[i] == '/' {
			return symbol[i+1:]
		}
	}
	return symbol
}
