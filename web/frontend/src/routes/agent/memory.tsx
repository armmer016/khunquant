import { createFileRoute } from "@tanstack/react-router"

import { AgentMemoryPage } from "@/components/agent-memory/agent-memory-page"

export const Route = createFileRoute("/agent/memory")({
  component: AgentMemoryRoute,
})

function AgentMemoryRoute() {
  return <AgentMemoryPage />
}
