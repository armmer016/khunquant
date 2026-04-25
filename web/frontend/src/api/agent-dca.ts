export interface DCAPlanListItem {
  id: number
  name: string
  provider: string
  account: string
  symbol: string
  amount_per_order: number
  frequency_expr: string
  timezone: string
  start_date: string
  end_date?: string
  enabled: boolean
  total_invested: number
  total_quantity: number
  avg_cost: number
  created_at: string
  updated_at: string
}

export interface DCAExecutionItem {
  id: number
  plan_id: number
  executed_at: string
  symbol: string
  provider: string
  account: string
  order_id: string
  amount_quote: number
  filled_price: number
  filled_quantity: number
  fee_quote: number
  status: "completed" | "failed"
  error_msg?: string
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, options)
  if (!res.ok) {
    let message = `API error: ${res.status} ${res.statusText}`
    try {
      const text = await res.text()
      if (text.trim()) message = text.trim()
    } catch {
      // ignore
    }
    throw new Error(message)
  }
  return res.json() as Promise<T>
}

export interface ListDCAPlansParams {
  enabled?: boolean
}

export async function listDCAPlans(params?: ListDCAPlansParams): Promise<DCAPlanListItem[]> {
  const q = new URLSearchParams()
  if (params?.enabled !== undefined) q.set("enabled", String(params.enabled))
  const qs = q.toString()
  return request<DCAPlanListItem[]>(`/api/agent/dca/plans${qs ? `?${qs}` : ""}`)
}

export async function getDCAPlan(id: number): Promise<DCAPlanListItem> {
  return request<DCAPlanListItem>(`/api/agent/dca/plans/${id}`)
}

export interface ListDCAExecutionsParams {
  limit?: number
  offset?: number
}

export async function getDCAExecutions(
  planId: number,
  params?: ListDCAExecutionsParams,
): Promise<DCAExecutionItem[]> {
  const q = new URLSearchParams()
  if (params?.limit) q.set("limit", String(params.limit))
  if (params?.offset) q.set("offset", String(params.offset))
  const qs = q.toString()
  return request<DCAExecutionItem[]>(`/api/agent/dca/plans/${planId}/executions${qs ? `?${qs}` : ""}`)
}
