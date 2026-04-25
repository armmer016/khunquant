package dca

import "time"

// TriggerConfig defines indicator-based execution conditions.
// Stored as JSON in the dca_plans.trigger_config column.
// When TriggerConfig is nil the plan runs on its fixed cron schedule (legacy behaviour).
type TriggerConfig struct {
	// Indicator and data settings
	Indicator string `json:"indicator"`       // "sma","ema","rsi","macd","bb","atr","stoch","vwap"
	Timeframe string `json:"timeframe"`       // "1m","5m","15m","1h","4h","1d","1w"
	Condition string `json:"condition"`       // condition identifier — see table in dca_condition.go
	Limit     int    `json:"limit,omitempty"` // OHLCV bars to fetch (default 100)

	// Indicator parameters (all optional; sensible defaults applied at runtime)
	Period     int     `json:"period,omitempty"`     // primary period (RSI=14, SMA/EMA/BB=20, ATR=14, Stoch K=14)
	Period2    int     `json:"period2,omitempty"`    // secondary period (MACD slow=26, Stoch D=3, EMA slow=50)
	Period3    int     `json:"period3,omitempty"`    // tertiary period (MACD signal=9)
	Multiplier float64 `json:"multiplier,omitempty"` // BB std-dev multiplier (default 2.0)
	Threshold  float64 `json:"threshold,omitempty"`  // custom comparison level (RSI, ATR, Stoch)
}

// Plan represents a DCA (Dollar Cost Averaging) plan configuration.
type Plan struct {
	ID             int64
	Name           string
	Provider       string
	Account        string
	Symbol         string
	AmountPerOrder float64 // quote currency amount to spend/receive per execution (e.g. 100 USDT)
	FrequencyExpr  string  // cron expression: "0 9 * * 1" = every Monday at 9am
	Timezone       string
	CronJobID      string     // ID of the associated cron service job
	StartDate      time.Time
	EndDate        *time.Time // nil = ongoing
	Enabled        bool
	TotalInvested  float64 // cumulative quote spent across all executions
	TotalQuantity  float64 // cumulative base asset acquired/sold
	AvgCost        float64 // volume-weighted average purchase price

	// Extended fields (added in migration)
	Side             string         // "buy" (default) | "sell"
	TriggerConfig    *TriggerConfig // nil = schedule-based (original behaviour)
	MaxExecPerPeriod int            // 0 = unlimited
	ExecPeriod       string         // "hour" | "day" | "week" | "" (no guardrail)
	NotifyChannel    string         // channel to route cron results to (e.g. "telegram")
	NotifyChatID     string         // chatID / user ID to route cron results to

	CreatedAt time.Time
	UpdatedAt time.Time
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
	AmountQuote    float64 // quote currency spent (buy) or received (sell)
	FilledPrice    float64 // execution price
	FilledQuantity float64 // base asset amount
	FeeQuote       float64 // fee in quote currency
	Status         string  // "completed" | "failed" | "skipped"
	ErrorMsg       string
	CreatedAt      time.Time
}

// DCASummary holds PnL and aggregate statistics for a plan.
type DCASummary struct {
	PlanID           int64
	TotalInvested    float64
	TotalQuantity    float64
	AvgCost          float64
	CurrentPrice     float64
	CurrentValue     float64 // total_quantity * current_price
	UnrealizedPnL    float64 // current_value - total_invested
	UnrealizedPnLPct float64 // unrealized_pnl / total_invested * 100
	ExecutionCount   int
	LastExecutionAt  *time.Time
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
