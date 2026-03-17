import { createFileRoute } from "@tanstack/react-router"

import { AgentConfigPage } from "@/components/agent-config/agent-config-page"

export const Route = createFileRoute("/agent/config")({
  component: AgentConfigRoute,
})

function AgentConfigRoute() {
  return <AgentConfigPage />
}
