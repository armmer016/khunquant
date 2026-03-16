import { createFileRoute } from "@tanstack/react-router"

import { PairingPage } from "@/components/pairing/pairing-page"

export const Route = createFileRoute("/agent/pairing")({
  component: PairingRoute,
})

function PairingRoute() {
  return <PairingPage />
}
