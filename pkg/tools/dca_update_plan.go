package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/adhocore/gronx"
	"github.com/cryptoquantumwave/khunquant/pkg/cron"
	"github.com/cryptoquantumwave/khunquant/pkg/dca"
)

// UpdateDCAPlanTool updates a DCA plan's enabled state, schedule, or end date.
type UpdateDCAPlanTool struct {
	store       *dca.Store
	cronService *cron.CronService
}

func NewUpdateDCAPlanTool(store *dca.Store, cronService *cron.CronService) *UpdateDCAPlanTool {
	return &UpdateDCAPlanTool{store: store, cronService: cronService}
}

func (t *UpdateDCAPlanTool) Name() string { return NameUpdateDCAPlan }

func (t *UpdateDCAPlanTool) Description() string {
	return "Update an existing DCA plan. You can pause/resume it (enabled), change the cron schedule (frequency_expr), or set an end date. If the schedule changes, the cron job is recreated automatically."
}

func (t *UpdateDCAPlanTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"plan_id": map[string]any{"type": "integer", "description": "ID of the plan to update."},
			"enabled": map[string]any{"type": "boolean", "description": "Enable or disable the plan."},
			"frequency_expr": map[string]any{
				"type":        "string",
				"description": "New cron expression (e.g. '0 12 * * 0' for Sundays at noon). Recreates the cron job.",
			},
			"timezone": map[string]any{
				"type":        "string",
				"description": "New timezone for the cron expression. Only applied when frequency_expr is also provided.",
			},
			"end_date": map[string]any{
				"type":        "string",
				"description": "New end date (YYYY-MM-DD) or 'none' to remove the end date.",
			},
		},
		"required": []string{"plan_id"},
	}
}

func (t *UpdateDCAPlanTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	planIDf, _ := args["plan_id"].(float64)
	planID := int64(planIDf)
	if planID <= 0 {
		return ErrorResult("plan_id is required")
	}

	plan, err := t.store.GetPlan(ctx, planID)
	if err != nil {
		return ErrorResult(fmt.Sprintf("plan not found: %v", err))
	}

	changed := false

	if v, ok := args["enabled"].(bool); ok {
		plan.Enabled = v
		t.cronService.EnableJob(plan.CronJobID, v)
		changed = true
	}

	if newExpr, ok := args["frequency_expr"].(string); ok && newExpr != "" {
		gx := gronx.New()
		if !gx.IsValid(newExpr) {
			return ErrorResult(fmt.Sprintf("invalid cron expression %q", newExpr))
		}

		if newTZ, ok := args["timezone"].(string); ok && newTZ != "" {
			plan.Timezone = newTZ
		}

		t.cronService.RemoveJob(plan.CronJobID)

		cronMsg := fmt.Sprintf("[DCA-AUTO] Execute plan: %s plan_id=%d", plan.Name, plan.ID)
		job, err := t.cronService.AddJob(
			fmt.Sprintf("dca:%d:%s", plan.ID, plan.Name),
			cron.CronSchedule{Kind: "cron", Expr: newExpr, TZ: plan.Timezone},
			cronMsg,
			false,
			"", "",
		)
		if err != nil {
			return ErrorResult(fmt.Sprintf("failed to recreate cron job: %v", err))
		}
		plan.FrequencyExpr = newExpr
		plan.CronJobID = job.ID
		changed = true
	}

	if endDateStr, ok := args["end_date"].(string); ok {
		if endDateStr == "none" || endDateStr == "" {
			plan.EndDate = nil
		} else {
			parsed, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				parsed, err = time.Parse(time.RFC3339, endDateStr)
				if err != nil {
					return ErrorResult(fmt.Sprintf("invalid end_date %q — use YYYY-MM-DD", endDateStr))
				}
			}
			plan.EndDate = &parsed
		}
		changed = true
	}

	if !changed {
		return UserResult("No changes specified. Provide at least one of: enabled, frequency_expr, end_date.")
	}

	if err := t.store.UpdatePlan(ctx, plan); err != nil {
		return ErrorResult(fmt.Sprintf("failed to update plan: %v", err))
	}

	status := "enabled"
	if !plan.Enabled {
		status = "disabled"
	}
	return UserResult(fmt.Sprintf("Plan %d (%s) updated: %s, schedule=%s (%s)\n",
		plan.ID, plan.Name, status, plan.FrequencyExpr, plan.Timezone))
}
