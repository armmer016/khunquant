---
name: dca
description: Set up, manage, and automate Dollar Cost Averaging (DCA) plans. Automatically execute periodic market buy orders on any exchange at a fixed quote amount per interval.
---

# Dollar Cost Averaging (DCA)

Systematically accumulate assets over time with automated, pre-authorized periodic purchases.

## Workflow

```
User asks to set up DCA
  → show plan summary, ask for confirmation
  → create_dca_plan (stores plan + schedules cron job)

Cron fires on schedule:
  → agent receives: "[DCA-AUTO] Execute plan: <name> plan_id=<id>"
  → call execute_dca_order with plan_id (no extra confirmation needed)
  → execution recorded in DCA snapshot database

User reviews performance:
  → get_dca_summary (avg cost, PnL%)
  → get_dca_history (detailed execution log)
```

## Automated Execution

When you receive a message starting with `[DCA-AUTO]`, this is a cron-triggered DCA task:

1. Extract the `plan_id` from the message (e.g., `plan_id=3`)
2. **Immediately call `execute_dca_order` with that plan_id** — no additional user confirmation is required
3. The plan was already user-approved at creation time; `execute_dca_order` is pre-authorized for this purpose

Example trigger message:
```
[DCA-AUTO] Execute plan: BTC Weekly plan_id=3
```

Response:
```
execute_dca_order({"plan_id": 3})
```

## Setting Up a Plan

Always show the user the plan summary and ask "Ready to activate?" before calling `create_dca_plan`.

```
create_dca_plan({
  "plan_name": "BTC Weekly",
  "provider": "binance",
  "account": "",
  "symbol": "BTC/USDT",
  "amount_per_order": 100,
  "frequency_expr": "0 9 * * 1",
  "timezone": "Asia/Bangkok",
  "start_date": "2026-05-01"
})
```

Returns the plan ID and cron job ID confirming the schedule is active.

## Managing Plans

**List all plans** (with stats):
```
list_dca_plans({})
list_dca_plans({"filter_enabled": true})
```

**Pause a plan** (cron job stays but skips execution):
```
update_dca_plan({"plan_id": 3, "enabled": false})
```

**Resume a paused plan**:
```
update_dca_plan({"plan_id": 3, "enabled": true})
```

**Change schedule** (recreates cron job automatically):
```
update_dca_plan({"plan_id": 3, "frequency_expr": "0 12 * * 0", "timezone": "Asia/Bangkok"})
```

**Set expiry** (plan auto-stops after this date):
```
update_dca_plan({"plan_id": 3, "end_date": "2027-01-01"})
```

**Delete plan** (removes cron job and all history — irreversible):
```
delete_dca_plan({"plan_id": 3})
```

Use `update_dca_plan enabled=false` instead of delete if you might want to resume.

## Reviewing Performance

**PnL summary** (fetches live price for unrealized PnL):
```
get_dca_summary({"plan_id": 3})
```

Returns: total invested, quantity acquired, avg cost (VWAP), current value, unrealized PnL (%).

**Execution log**:
```
get_dca_history({"plan_id": 3, "limit": 20})
```

## Common Cron Schedules

| Schedule | Expression |
|----------|-----------|
| Daily at 9am | `0 9 * * *` |
| Every Monday 9am | `0 9 * * 1` |
| Mon & Thu 10am | `0 10 * * 1,4` |
| 1st of month 9am | `0 9 1 * *` |

All times are in the specified `timezone` (default: UTC).

## Safety Rules

1. **Confirm before creating** — Always show the plan details (asset, amount, frequency) and ask the user to confirm before calling `create_dca_plan`.
2. **No double-confirmation for cron** — `[DCA-AUTO]` messages are pre-authorized; call `execute_dca_order` directly.
3. **Market orders only** — All DCA executions are market buys (immediate fill at best price).
4. **Failure is non-fatal** — If an execution fails (low balance, rate limit, etc.), it is recorded with `status=failed` and the plan stays active for the next window. The agent should notify the user if this happens.
5. **Rate limits** — Each execution places one market order. If running multiple plans, check `get_order_rate_status` first.
6. **Minimum order** — Ensure `amount_per_order` is at or above the exchange minimum (typically ~$10 USDT).
7. **Irreversible trades** — DCA orders are live market orders. `delete_dca_plan` only removes future executions; past trades cannot be undone.
