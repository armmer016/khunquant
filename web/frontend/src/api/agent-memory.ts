export interface AgentMemoryFile {
  name: string
  path: string // relative to memory dir, e.g. "MEMORY.md" or "202603/20260316.md"
  size: number
  modified_at: string
}

export interface AgentMemoryFileContent {
  path: string
  content: string
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

function encodePath(path: string): string {
  // Encode each segment separately to preserve slashes as path separators
  return path
    .split("/")
    .map((seg) => encodeURIComponent(seg))
    .join("/")
}

export async function getMemoryFiles(): Promise<AgentMemoryFile[]> {
  return request<AgentMemoryFile[]>("/api/agent/memory/files")
}

export async function getMemoryFile(
  path: string,
): Promise<AgentMemoryFileContent> {
  return request<AgentMemoryFileContent>(
    `/api/agent/memory/files/${encodePath(path)}`,
  )
}

export async function saveMemoryFile(
  path: string,
  content: string,
): Promise<void> {
  await request(`/api/agent/memory/files/${encodePath(path)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ path, content }),
  })
}

export async function createMemoryFile(
  path: string,
  content: string,
): Promise<{ path: string }> {
  return request<{ path: string }>("/api/agent/memory/files", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ path, content }),
  })
}

export async function deleteMemoryFile(path: string): Promise<void> {
  await request(`/api/agent/memory/files/${encodePath(path)}`, {
    method: "DELETE",
  })
}

export interface MemorySize {
  general_bytes: number
  snapshot_bytes: number
  dca_bytes: number
  total_bytes: number
}

export async function getMemorySize(): Promise<MemorySize> {
  return request<MemorySize>("/api/agent/memory/size")
}
