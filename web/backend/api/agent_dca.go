package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cryptoquantumwave/khunquant/pkg/config"
	"github.com/cryptoquantumwave/khunquant/pkg/dca"
)

type dcaPlanListItem struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Provider       string     `json:"provider"`
	Account        string     `json:"account"`
	Symbol         string     `json:"symbol"`
	AmountPerOrder float64    `json:"amount_per_order"`
	FrequencyExpr  string     `json:"frequency_expr"`
	Timezone       string     `json:"timezone"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	Enabled        bool       `json:"enabled"`
	TotalInvested  float64    `json:"total_invested"`
	TotalQuantity  float64    `json:"total_quantity"`
	AvgCost        float64    `json:"avg_cost"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type dcaExecutionItem struct {
	ID             int64     `json:"id"`
	PlanID         int64     `json:"plan_id"`
	ExecutedAt     time.Time `json:"executed_at"`
	Symbol         string    `json:"symbol"`
	Provider       string    `json:"provider"`
	Account        string    `json:"account"`
	OrderID        string    `json:"order_id"`
	AmountQuote    float64   `json:"amount_quote"`
	FilledPrice    float64   `json:"filled_price"`
	FilledQuantity float64   `json:"filled_quantity"`
	FeeQuote       float64   `json:"fee_quote"`
	Status         string    `json:"status"`
	ErrorMsg       string    `json:"error_msg,omitempty"`
}

func (h *Handler) registerAgentDCARoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/agent/dca/plans", h.handleListDCAPlans)
	mux.HandleFunc("GET /api/agent/dca/plans/{id}", h.handleGetDCAPlan)
	mux.HandleFunc("GET /api/agent/dca/plans/{id}/executions", h.handleGetDCAExecutions)
}

func (h *Handler) dcaWorkspacePath() (string, error) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	return cfg.WorkspacePath(), nil
}

func (h *Handler) handleListDCAPlans(w http.ResponseWriter, r *http.Request) {
	workspacePath, err := h.dcaWorkspacePath()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	store, err := dca.NewStore(workspacePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to open DCA store: %v", err), http.StatusInternalServerError)
		return
	}
	defer store.Close()

	q := r.URL.Query()
	var filterEnabled *bool
	if v := q.Get("enabled"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			filterEnabled = &b
		}
	}

	plans, err := store.ListPlans(r.Context(), dca.QueryFilter{Enabled: filterEnabled})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list DCA plans: %v", err), http.StatusInternalServerError)
		return
	}

	items := make([]dcaPlanListItem, len(plans))
	for i := range plans {
		items[i] = planToItem(&plans[i])
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items) //nolint:errcheck
}

func (h *Handler) handleGetDCAPlan(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	workspacePath, err := h.dcaWorkspacePath()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	store, err := dca.NewStore(workspacePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to open DCA store: %v", err), http.StatusInternalServerError)
		return
	}
	defer store.Close()

	plan, err := store.GetPlan(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("plan not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(planToItem(plan)) //nolint:errcheck
}

func (h *Handler) handleGetDCAExecutions(w http.ResponseWriter, r *http.Request) {
	planID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	workspacePath, err := h.dcaWorkspacePath()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	store, err := dca.NewStore(workspacePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to open DCA store: %v", err), http.StatusInternalServerError)
		return
	}
	defer store.Close()

	execs, err := store.GetExecutions(r.Context(), dca.QueryFilter{PlanID: &planID, Limit: limit, Offset: offset})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get executions: %v", err), http.StatusInternalServerError)
		return
	}

	items := make([]dcaExecutionItem, len(execs))
	for i, e := range execs {
		items[i] = dcaExecutionItem{
			ID:             e.ID,
			PlanID:         e.PlanID,
			ExecutedAt:     e.ExecutedAt,
			Symbol:         e.Symbol,
			Provider:       e.Provider,
			Account:        e.Account,
			OrderID:        e.OrderID,
			AmountQuote:    e.AmountQuote,
			FilledPrice:    e.FilledPrice,
			FilledQuantity: e.FilledQuantity,
			FeeQuote:       e.FeeQuote,
			Status:         e.Status,
			ErrorMsg:       e.ErrorMsg,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items) //nolint:errcheck
}

func planToItem(p *dca.Plan) dcaPlanListItem {
	return dcaPlanListItem{
		ID:             p.ID,
		Name:           p.Name,
		Provider:       p.Provider,
		Account:        p.Account,
		Symbol:         p.Symbol,
		AmountPerOrder: p.AmountPerOrder,
		FrequencyExpr:  p.FrequencyExpr,
		Timezone:       p.Timezone,
		StartDate:      p.StartDate,
		EndDate:        p.EndDate,
		Enabled:        p.Enabled,
		TotalInvested:  p.TotalInvested,
		TotalQuantity:  p.TotalQuantity,
		AvgCost:        p.AvgCost,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
