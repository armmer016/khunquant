export interface PairingRequest {
  code: string
  platform: string
  platform_id: string
  username?: string
  display_name?: string
  canonical_id: string
  chat_id: number
  created_at_ms: number
  expires_at_ms: number
}

export async function getPairingRequests(): Promise<PairingRequest[]> {
  const res = await fetch("/api/pairing/requests")
  if (!res.ok) throw new Error(await res.text())
  const data = await res.json()
  return data.requests ?? []
}

export async function approvePairing(code: string): Promise<void> {
  const res = await fetch(`/api/pairing/approve/${code}`, { method: "POST" })
  if (!res.ok) throw new Error(await res.text())
}

export async function rejectPairing(code: string): Promise<void> {
  const res = await fetch(`/api/pairing/reject/${code}`, { method: "POST" })
  if (!res.ok) throw new Error(await res.text())
}
