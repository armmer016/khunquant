import { IconCheck, IconLoader2, IconShieldLock, IconX } from "@tabler/icons-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  type PairingRequest,
  approvePairing,
  getPairingRequests,
  rejectPairing,
} from "@/api/pairing"
import { PageHeader } from "@/components/page-header"
import { updateGatewayStore } from "@/store"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"

function formatExpiry(expiresAtMs: number): string {
  const remaining = expiresAtMs - Date.now()
  if (remaining <= 0) return "Expired"
  const minutes = Math.ceil(remaining / 60_000)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  return mins > 0 ? `${hours}h ${mins}m` : `${hours}h`
}

function PairingCard({
  req,
  onApprove,
  onReject,
  approving,
  rejecting,
}: {
  req: PairingRequest
  onApprove: () => void
  onReject: () => void
  approving: boolean
  rejecting: boolean
}) {
  const { t } = useTranslation()
  const displayName = req.display_name || req.username || req.platform_id
  const username = req.username ? `@${req.username}` : null
  const expiry = formatExpiry(req.expires_at_ms)

  return (
    <Card className="border-border/60">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <CardTitle className="text-sm font-semibold">{displayName}</CardTitle>
            {username && (
              <CardDescription className="text-xs">{username}</CardDescription>
            )}
          </div>
          <span className="bg-muted text-muted-foreground shrink-0 rounded px-2 py-0.5 font-mono text-xs">
            {req.code}
          </span>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="text-muted-foreground mb-4 flex flex-wrap items-center gap-3 text-xs">
          <span>
            {t("pages.agent.pairing.platform")}: {req.platform}
          </span>
          <span>
            {t("pages.agent.pairing.id")}: {req.platform_id}
          </span>
          <span>
            {t("pages.agent.pairing.expires")}: {expiry}
          </span>
        </div>
        <div className="flex gap-2">
          <Button
            size="sm"
            onClick={onApprove}
            disabled={approving || rejecting}
            className="gap-1.5"
          >
            {approving ? (
              <IconLoader2 className="size-3.5 animate-spin" />
            ) : (
              <IconCheck className="size-3.5" />
            )}
            {t("pages.agent.pairing.approve")}
          </Button>
          <Button
            size="sm"
            variant="outline"
            onClick={onReject}
            disabled={approving || rejecting}
            className="gap-1.5"
          >
            {rejecting ? (
              <IconLoader2 className="size-3.5 animate-spin" />
            ) : (
              <IconX className="size-3.5" />
            )}
            {t("pages.agent.pairing.reject")}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

export function PairingPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const {
    data: requests,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["pairing-requests"],
    queryFn: getPairingRequests,
    refetchInterval: 5_000,
  })

  const approveMutation = useMutation({
    mutationFn: approvePairing,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["pairing-requests"] })
      updateGatewayStore({ restartRequired: true })
      toast.success(t("pages.agent.pairing.approve_success"))
    },
    onError: (err: Error) => {
      toast.error(t("pages.agent.pairing.approve_error"), { description: err.message })
    },
  })

  const rejectMutation = useMutation({
    mutationFn: rejectPairing,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["pairing-requests"] })
      toast.success(t("pages.agent.pairing.reject_success"))
    },
    onError: (err: Error) => {
      toast.error(t("pages.agent.pairing.reject_error"), { description: err.message })
    },
  })

  return (
    <div className="flex h-full flex-col">
      <PageHeader title={t("navigation.pairing")} />

      <div className="flex-1 overflow-auto px-6 py-3">
        <div className="w-full max-w-xl space-y-4">
          {isLoading ? (
            <div className="text-muted-foreground flex items-center gap-2 py-6 text-sm">
              <IconLoader2 className="size-4 animate-spin" />
              {t("pages.agent.pairing.loading")}
            </div>
          ) : error ? (
            <div className="text-destructive py-6 text-sm">
              {t("pages.agent.pairing.load_error")}
            </div>
          ) : !requests?.length ? (
            <Card className="border-dashed">
              <CardContent className="text-muted-foreground py-12 text-center text-sm">
                <IconShieldLock className="mx-auto mb-3 size-8 opacity-30" />
                <p>{t("pages.agent.pairing.empty")}</p>
                <p className="mt-1 text-xs opacity-70">
                  {t("pages.agent.pairing.empty_hint")}
                </p>
              </CardContent>
            </Card>
          ) : (
            <>
              <p className="text-muted-foreground text-sm">
                {t("pages.agent.pairing.description").replace(
                  "{{count}}",
                  String(requests.length),
                )}
              </p>
              <div className="space-y-3">
                {requests.map((req: PairingRequest) => (
                  <PairingCard
                    key={req.code}
                    req={req}
                    onApprove={() => approveMutation.mutate(req.code)}
                    onReject={() => rejectMutation.mutate(req.code)}
                    approving={
                      approveMutation.isPending &&
                      approveMutation.variables === req.code
                    }
                    rejecting={
                      rejectMutation.isPending &&
                      rejectMutation.variables === req.code
                    }
                  />
                ))}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
