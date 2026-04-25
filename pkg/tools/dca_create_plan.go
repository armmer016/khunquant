package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/adhocore/gronx"
	"github.com/cryptoquantumwave/khunquant/pkg/config"
	"github.com/cryptoquantumwave/khunquant/pkg/cron"
	"github.com/cryptoquantumwave/khunquant/pkg/dca"
)

// CreateDCAPlanTool creates a new DCA plan and schedules the recurring cron job.
type CreateDCAPlanTool struct {
	cfg         *config.Config
	store       *dca.Store
	cronService *cron.CronService
}

func NewCreateDCAPlanTool(cfg *config.Config, store *dca.Store, cronService *cron.CronService) *CreateDCAPlanTool {
	return &CreateDCAPlanTool{cfg: cfg, store: store, cronService: cronService}
}

func (t *CreateDCAPlanTool) Name() string { return NameCreateDCAPlan }

func (t *CreateDCAPlanTool) Description() string {
	return "Create a new Dollar Cost Averaging (DCA) plan that automatically purchases a fixed quote amount of an asset on a recurring schedule. After confirming the plan details with the user, this tool creates the plan and schedules the cron job."
}

func (t *CreateDCAPlanTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"plan_name":        map[string]any{"type": "string", "description": "Human-readable plan name (e.g. 'BTC Weekly')."},
			"provider":         map[string]any{"type": "string", "description": "Exchange/provider name (e.g. 'binance')."},
			"account":          map[string]any{"type": "string", "description": "Account name (empty = default)."},
			"symbol":           map[string]any{"type": "string", "description": "Trading pair in CCXT format (e.g. 'BTC/USDT')."},
			"amount_per_order": map[string]any{"type": "number", "description": "Quote currency amount to spend per execution (e.g. 100 for 100 USDT)."},
			"frequency_expr": map[string]any{
				"type":        "string",
				"description": "Cron expression for the schedule (e.g. '0 9 * * 1' for every Monday at 9am UTC).",
			},
			"timezone": map[string]any{
				"type":        "string",
				"description": "IANA timezone for the cron expression (e.g. 'America/New_York', 'Asia/Bangkok'). Defaults to UTC.",
			},
			"start_date": map[string]any{
				"type":        "string",
				"description": "ISO 8601 date when the plan becomes active (e.g. '2026-05-01'). Defaults to today.",
			},
			"end_date": map[string]any{
				"type":        "string",
				"description": "Optional ISO 8601 date when the plan stops (e.g. '2027-01-01'). Omit for ongoing.",
			},
		},
		"required": []string{"plan_name", "provider", "symbol", "amount_per_order", "frequency_expr"},
	}
}

func (t *CreateDCAPlanTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	planName, _ := args["plan_name"].(string)
	providerID, _ := args["provider"].(string)
	account, _ := args["account"].(string)
	symbol, _ := args["symbol"].(string)
	amountPerOrder, _ := args["amount_per_order"].(float64)
	frequencyExpr, _ := args["frequency_expr"].(string)
	timezone, _ := args["timezone"].(string)
	startDateStr, _ := args["start_date"].(string)
	endDateStr, _ := args["end_date"].(string)

	if planName == "" || providerID == "" || symbol == "" {
		return ErrorResult("plan_name, provider, and symbol are required")
	}
	if amountPerOrder <= 0 {
		return ErrorResult("amount_per_order must be positive")
	}
	if frequencyExpr == "" {
		return ErrorResult("frequency_expr is required (cron expression)")
	}

	gx := gronx.New()
	if !gx.IsValid(frequencyExpr) {
		return ErrorResult(fmt.Sprintf("invalid cron expression %q — use standard 5-field cron format (e.g. '0 9 * * 1')", frequencyExpr))
	}

	if timezone == "" {
		timezone = "UTC"
	}

	startDate := time.Now().UTC()
	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, startDateStr)
			if err != nil {
				return ErrorResult(fmt.Sprintf("invalid start_date %q — use YYYY-MM-DD format", startDateStr))
			}
		}
		startDate = parsed
	}

	var endDate *time.Time
	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			parsed, err = time.Parse(time.RFC3339, endDateStr)
			if err != nil {
				return ErrorResult(fmt.Sprintf("invalid end_date %q — use YYYY-MM-DD format", endDateStr))
			}
		}
		endDate = &parsed
	}

	now := time.Now().UTC()
	plan := &dca.Plan{
		Name:           planName,
		Provider:       providerID,
		Account:        account,
		Symbol:         symbol,
		AmountPerOrder: amountPerOrder,
		FrequencyExpr:  frequencyExpr,
		Timezone:       timezone,
		StartDate:      startDate,
		EndDate:        endDate,
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	planID, err := t.store.SavePlan(ctx, plan)
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to save DCA plan: %v", err))
	}

	cronMsg := fmt.Sprintf("[DCA-AUTO] Execute plan: %s plan_id=%d", planName, planID)
	job, err := t.cronService.AddJob(
		fmt.Sprintf("dca:%d:%s", planID, planName),
		cron.CronSchedule{Kind: "cron", Expr: frequencyExpr, TZ: timezone},
		cronMsg,
		false,
		"", "",
	)
	if err != nil {
		_ = t.store.DeletePlan(ctx, planID)
		return ErrorResult(fmt.Sprintf("failed to schedule cron job: %v", err))
	}

	plan.CronJobID = job.ID
	if err := t.store.UpdatePlan(ctx, plan); err != nil {
		return ErrorResult(fmt.Sprintf("failed to update plan with cron job ID: %v", err))
	}

	out := fmt.Sprintf("DCA plan created successfully!\n\n")
	out += fmt.Sprintf("  Plan ID:       %d\n", planID)
	out += fmt.Sprintf("  Name:          %s\n", planName)
	out += fmt.Sprintf("  Symbol:        %s on %s\n", symbol, providerID)
	out += fmt.Sprintf("  Amount/order:  %.2f (quote currency)\n", amountPerOrder)
	out += fmt.Sprintf("  Schedule:      %s (%s)\n", frequencyExpr, timezone)
	out += fmt.Sprintf("  Cron job ID:   %s\n", job.ID)
	out += fmt.Sprintf("\nThe plan is now active. On each scheduled trigger, the agent will receive:\n")
	out += fmt.Sprintf("  \"%s\"\n", cronMsg)
	out += fmt.Sprintf("\nThe DCA skill will then call execute_dca_order with plan_id=%d automatically.\n", planID)
	return UserResult(out)
}
