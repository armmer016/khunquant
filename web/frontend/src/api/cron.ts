export interface CronSchedule {
  kind: "at" | "every" | "cron"
  atMs?: number
  everyMs?: number
  expr?: string
  tz?: string
}

export interface CronPayload {
  kind: string
  message: string
  command?: string
  deliver: boolean
  channel?: string
  to?: string
}

export interface CronJobState {
  nextRunAtMs?: number
  lastRunAtMs?: number
  lastStatus?: string
  lastError?: string
}

export interface CronJob {
  id: string
  name: string
  enabled: boolean
  schedule: CronSchedule
  payload: CronPayload
  state: CronJobState
  createdAtMs: number
  updatedAtMs: number
  deleteAfterRun: boolean
}

interface CronJobsResponse {
  jobs: CronJob[]
}

interface CronActionResponse {
  status: string
}

export interface CronUpdateRequest {
  name?: string
  message?: string
  enabled?: boolean
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, options)
  if (!res.ok) {
    let message = `API error: ${res.status} ${res.statusText}`
    try {
      const body = (await res.json()) as { error?: string }
      if (typeof body.error === "string" && body.error.trim() !== "") {
        message = body.error
      }
    } catch {
      // ignore
    }
    throw new Error(message)
  }
  return res.json() as Promise<T>
}

export async function getCronJobs(): Promise<CronJob[]> {
  const res = await request<CronJobsResponse>("/api/cron/jobs")
  return res.jobs ?? []
}

export async function updateCronJob(
  id: string,
  patch: CronUpdateRequest,
): Promise<CronActionResponse> {
  return request<CronActionResponse>(
    `/api/cron/jobs/${encodeURIComponent(id)}`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(patch),
    },
  )
}

export async function deleteCronJob(id: string): Promise<CronActionResponse> {
  return request<CronActionResponse>(
    `/api/cron/jobs/${encodeURIComponent(id)}`,
    { method: "DELETE" },
  )
}
