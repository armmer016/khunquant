export interface SnapshotListItem {
  id: number
  taken_at: string
  quote: string
  total_value: number
  label: string
  note: string
}

export interface SnapshotPosition {
  source: string
  account: string
  category: string
  asset: string
  quantity: number
  quote: string
  price: number
  value: number
  meta?: Record<string, string>
}

export interface SnapshotDetail extends SnapshotListItem {
  positions: SnapshotPosition[]
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

export interface ListSnapshotsParams {
  limit?: number
  offset?: number
  label?: string
}

export async function listSnapshots(
  params?: ListSnapshotsParams,
): Promise<SnapshotListItem[]> {
  const q = new URLSearchParams()
  if (params?.limit) q.set("limit", String(params.limit))
  if (params?.offset) q.set("offset", String(params.offset))
  if (params?.label) q.set("label", params.label)
  const qs = q.toString()
  return request<SnapshotListItem[]>(`/api/agent/snapshots${qs ? `?${qs}` : ""}`)
}

export async function getSnapshot(id: number): Promise<SnapshotDetail> {
  return request<SnapshotDetail>(`/api/agent/snapshots/${id}`)
}

export async function deleteSnapshot(id: number): Promise<void> {
  await request(`/api/agent/snapshots/${id}`, { method: "DELETE" })
}
