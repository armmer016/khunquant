export interface AgentConfigFile {
  name: string
  size: number
  modified_at: string
}

export interface AgentConfigFileContent {
  name: string
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

export async function getAgentConfigFiles(): Promise<AgentConfigFile[]> {
  return request<AgentConfigFile[]>("/api/agent/config/files")
}

export async function getAgentConfigFile(
  name: string,
): Promise<AgentConfigFileContent> {
  return request<AgentConfigFileContent>(
    `/api/agent/config/files/${encodeURIComponent(name)}`,
  )
}

export async function saveAgentConfigFile(
  name: string,
  content: string,
): Promise<void> {
  await request(`/api/agent/config/files/${encodeURIComponent(name)}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, content }),
  })
}

export async function createAgentConfigFile(
  name: string,
  content: string,
): Promise<{ name: string }> {
  return request<{ name: string }>("/api/agent/config/files", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, content }),
  })
}

export async function deleteAgentConfigFile(name: string): Promise<void> {
  await request(`/api/agent/config/files/${encodeURIComponent(name)}`, {
    method: "DELETE",
  })
}
