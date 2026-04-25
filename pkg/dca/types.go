package dca

import "time"

// Plan represents a DCA (Dollar Cost Averaging) plan configuration.
type Plan struct {
	ID             int64
	Name           string
	Provider       string
	Account        string
	Symbol         string
	AmountPerOrder float64 // quote currency amount to spend per execution (e.g. 100 USDT)
	FrequencyExpr  string  // cron expression: "0 9 * * 1" = every Monday at 9am
	Timezone       string
	CronJobID      string     // ID of the associated cron service job
	StartDate      time.Time
	EndDate        *time.Time // nil = ongoing
	Enabled        bool
	TotalInvested  float64 // cumulative quote spent across all executions
	TotalQuantity  float64 // cumulative base asset acquired
	AvgCost        float64 // volume-weighted average purchase price
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Execution records a single DCA order that was placed by the scheduler.
type Execution struct {
	ID             int64
	PlanID         int64
	ExecutedAt     time.Time
	Symbol         string
	Provider       string
	Account        string
	OrderID        string
	AmountQuote    float64 // quote currency spent
	FilledPrice    float64 // execution price
	FilledQuantity float64 // base asset received
	FeeQuote       float64 // fee in quote currency
	Status         string  // "completed" | "failed"
	ErrorMsg       string
	CreatedAt      time.Time
}

// DCASummary holds PnL and aggregate statistics for a plan.
type DCASummary struct {
	PlanID          int64
	TotalInvested   float64
	TotalQuantity   float64
	AvgCost         float64
	CurrentPrice    float64
	CurrentValue    float64 // total_quantity * current_price
	UnrealizedPnL   float64 // current_value - total_invested
	UnrealizedPnLPct float64 // unrealized_pnl / total_invested * 100
	ExecutionCount  int
	LastExecutionAt *time.Time
}

// QueryFilter controls which plans or executions are returned.
type QueryFilter struct {
	PlanID  *int64
	Enabled *bool
	Since   *time.Time
	Until   *time.Time
	Limit   int
	Offset  int
}
