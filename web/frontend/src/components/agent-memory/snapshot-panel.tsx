import { IconLoader2, IconTrash } from "@tabler/icons-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  type SnapshotListItem,
  deleteSnapshot,
  getSnapshot,
  listSnapshots,
} from "@/api/agent-snapshot"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

function formatValue(value: number, quote: string): string {
  return `${value.toLocaleString(undefined, { maximumFractionDigits: 2 })} ${quote}`
}

interface SnapshotPanelProps {
  onDeleteSuccess?: () => void
}

export function SnapshotPanel({ onDeleteSuccess }: SnapshotPanelProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  const { data: snapshots, isLoading: isListLoading } = useQuery({
    queryKey: ["agent-snapshots"],
    queryFn: () => listSnapshots({ limit: 100 }),
  })

  const { data: detail, isLoading: isDetailLoading } = useQuery({
    queryKey: ["agent-snapshot", selectedId],
    queryFn: () => getSnapshot(selectedId!),
    enabled: selectedId !== null,
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteSnapshot(selectedId!),
    onSuccess: () => {
      toast.success(t("pages.agent.agent_memory.snapshot.delete_success"))
      setShowDeleteDialog(false)
      setSelectedId(null)
      void queryClient.invalidateQueries({ queryKey: ["agent-snapshots"] })
      onDeleteSuccess?.()
    },
    onError: (err) => {
      toast.error(
        err instanceof Error
          ? err.message
          : t("pages.agent.agent_memory.snapshot.delete_error"),
      )
    },
  })

  const itemClass = (id: number) =>
    `w-full rounded-md px-3 py-2 text-left text-sm transition-colors ${
      selectedId === id
        ? "bg-accent/80 text-foreground font-medium"
        : "text-muted-foreground hover:bg-muted/60"
    }`

  const selectedSnapshot = snapshots?.find((s) => s.id === selectedId)

  return (
    <div className="flex min-h-0 flex-1 overflow-hidden">
      {/* Left panel: list */}
      <div className="border-border/40 flex w-64 shrink-0 flex-col border-r">
        <div className="flex-1 overflow-auto p-2">
          {isListLoading ? (
            <div className="text-muted-foreground p-2 text-sm">
              {t("labels.loading")}
            </div>
          ) : !snapshots || snapshots.length === 0 ? (
            <div className="text-muted-foreground p-2 text-sm">
              {t("pages.agent.agent_memory.snapshot.empty")}
            </div>
          ) : (
            <ul className="space-y-0.5">
              {snapshots.map((snap: SnapshotListItem) => (
                <li key={snap.id}>
                  <button
                    onClick={() => setSelectedId(snap.id)}
                    className={itemClass(snap.id)}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <span className="truncate text-xs font-mono">
                        {formatDate(snap.taken_at)}
                      </span>
                      {snap.label && (
                        <span className="bg-accent text-foreground shrink-0 rounded px-1.5 py-0.5 text-[10px]">
                          {snap.label}
                        </span>
                      )}
                    </div>
                    <div className="text-foreground/70 mt-0.5 text-xs">
                      {formatValue(snap.total_value, snap.quote)}
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Right panel: detail */}
      <div className="flex min-h-0 flex-1 flex-col overflow-auto p-4">
        {selectedId === null ? (
          <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
            {t("pages.agent.agent_memory.snapshot.select_hint")}
          </div>
        ) : isDetailLoading ? (
          <div className="flex h-full items-center justify-center">
            <IconLoader2 className="text-muted-foreground size-5 animate-spin" />
          </div>
        ) : detail ? (
          <div className="flex flex-col gap-4">
            {/* Header */}
            <div className="flex items-start justify-between gap-4">
              <div className="flex flex-col gap-1">
                <div className="text-foreground/90 font-mono text-sm font-medium">
                  {formatDate(detail.taken_at)}
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  {detail.label && (
                    <span className="bg-accent text-foreground rounded px-1.5 py-0.5 text-xs">
                      {detail.label}
                    </span>
                  )}
                  <span className="text-muted-foreground text-sm">
                    {t("pages.agent.agent_memory.snapshot.total_value")}:{" "}
                    <span className="text-foreground font-medium">
                      {formatValue(detail.total_value, detail.quote)}
                    </span>
                  </span>
                </div>
                {detail.note && (
                  <p className="text-muted-foreground text-sm">{detail.note}</p>
                )}
              </div>
              <Button
                variant="ghost"
                size="icon-sm"
                className="text-muted-foreground hover:text-destructive shrink-0"
                onClick={() => setShowDeleteDialog(true)}
                title={t("pages.agent.agent_memory.snapshot.delete_title")}
              >
                <IconTrash className="size-4" />
              </Button>
            </div>

            {/* Positions table */}
            {detail.positions && detail.positions.length > 0 && (
              <div className="overflow-hidden rounded-lg border">
                <div className="border-border/50 border-b px-3 py-2">
                  <span className="text-foreground/80 text-sm font-medium">
                    {t("pages.agent.agent_memory.snapshot.positions")}
                  </span>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="bg-muted/40 text-muted-foreground border-b text-xs uppercase tracking-wide">
                        <th className="px-3 py-2 text-left">
                          {t("pages.agent.agent_memory.snapshot.col_source")}
                        </th>
                        <th className="px-3 py-2 text-left">
                          {t("pages.agent.agent_memory.snapshot.col_account")}
                        </th>
                        <th className="px-3 py-2 text-left">
                          {t("pages.agent.agent_memory.snapshot.col_category")}
                        </th>
                        <th className="px-3 py-2 text-left">
                          {t("pages.agent.agent_memory.snapshot.col_asset")}
                        </th>
                        <th className="px-3 py-2 text-right">
                          {t("pages.agent.agent_memory.snapshot.col_quantity")}
                        </th>
                        <th className="px-3 py-2 text-right">
                          {t("pages.agent.agent_memory.snapshot.col_price")}
                        </th>
                        <th className="px-3 py-2 text-right">
                          {t("pages.agent.agent_memory.snapshot.col_value")}
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {detail.positions.map((p, i) => (
                        <tr
                          key={i}
                          className="border-border/30 border-b last:border-0"
                        >
                          <td className="px-3 py-2">{p.source}</td>
                          <td className="text-muted-foreground px-3 py-2">
                            {p.account || "—"}
                          </td>
                          <td className="text-muted-foreground px-3 py-2">
                            {p.category || "—"}
                          </td>
                          <td className="px-3 py-2 font-medium">{p.asset}</td>
                          <td className="px-3 py-2 text-right font-mono">
                            {p.quantity.toLocaleString(undefined, {
                              maximumFractionDigits: 8,
                            })}
                          </td>
                          <td className="px-3 py-2 text-right font-mono">
                            {p.price > 0
                              ? p.price.toLocaleString(undefined, {
                                  maximumFractionDigits: 4,
                                })
                              : "—"}
                            {p.quote && (
                              <span className="text-muted-foreground ml-1 text-xs">
                                {p.quote}
                              </span>
                            )}
                          </td>
                          <td className="px-3 py-2 text-right font-mono">
                            {p.value.toLocaleString(undefined, {
                              maximumFractionDigits: 2,
                            })}
                            {p.quote && (
                              <span className="text-muted-foreground ml-1 text-xs">
                                {p.quote}
                              </span>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        ) : null}
      </div>

      {/* Delete confirmation dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("pages.agent.agent_memory.snapshot.delete_title")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("pages.agent.agent_memory.snapshot.delete_description", {
                id: selectedId,
                date: selectedSnapshot ? formatDate(selectedSnapshot.taken_at) : "",
              })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMutation.isPending}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={deleteMutation.isPending}
              onClick={() => deleteMutation.mutate()}
            >
              {deleteMutation.isPending ? (
                <IconLoader2 className="size-4 animate-spin" />
              ) : (
                <IconTrash className="size-4" />
              )}
              {t("pages.agent.agent_memory.snapshot.delete_confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
