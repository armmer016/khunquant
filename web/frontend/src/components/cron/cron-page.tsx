import {
  IconClock,
  IconLoader2,
  IconPencil,
  IconPlayerPlay,
  IconPlayerStop,
  IconTrash,
} from "@tabler/icons-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import * as React from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  type CronJob,
  type CronUpdateRequest,
  deleteCronJob,
  getCronJobs,
  updateCronJob,
} from "@/api/cron"
import { PageHeader } from "@/components/page-header"
import { updateGatewayStore } from "@/store"
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
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Sheet,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/utils"

function formatSchedule(job: CronJob, t: (k: string) => string): string {
  const { schedule } = job
  if (schedule.kind === "every" && schedule.everyMs) {
    const secs = Math.round(schedule.everyMs / 1000)
    if (secs < 60)
      return t("pages.agent.cron.schedule_every_seconds").replace(
        "{{n}}",
        String(secs),
      )
    if (secs < 3600)
      return t("pages.agent.cron.schedule_every_minutes").replace(
        "{{n}}",
        String(Math.round(secs / 60)),
      )
    return t("pages.agent.cron.schedule_every_hours").replace(
      "{{n}}",
      String(Math.round(secs / 3600)),
    )
  }
  if (schedule.kind === "cron" && schedule.expr) return schedule.expr
  if (schedule.kind === "at" && schedule.atMs)
    return new Date(schedule.atMs).toLocaleString()
  return t("pages.agent.cron.schedule_unknown")
}

function formatNextRun(ms?: number, t?: (k: string) => string): string {
  if (!ms) return t ? t("pages.agent.cron.no_next_run") : "—"
  const diff = ms - Date.now()
  if (diff < 0) return t ? t("pages.agent.cron.overdue") : "overdue"
  if (diff < 60_000) return `${Math.round(diff / 1000)}s`
  if (diff < 3_600_000) return `${Math.round(diff / 60_000)}m`
  return new Date(ms).toLocaleTimeString()
}

function formatLastRun(ms?: number): string {
  if (!ms) return "—"
  return new Date(ms).toLocaleString()
}

// ── Edit sheet ────────────────────────────────────────────────────────────────

interface EditSheetProps {
  job: CronJob | null
  onClose: () => void
  onSave: (id: string, patch: CronUpdateRequest) => void
  saving: boolean
}

function EditSheet({ job, onClose, onSave, saving }: EditSheetProps) {
  const { t } = useTranslation()
  const [name, setName] = React.useState("")
  const [message, setMessage] = React.useState("")

  React.useEffect(() => {
    if (job) {
      setName(job.name)
      setMessage(job.payload.message)
    }
  }, [job])

  return (
    <Sheet open={!!job} onOpenChange={(open) => { if (!open) onClose() }}>
      <SheetContent side="right" className="flex flex-col sm:max-w-md">
        <SheetHeader>
          <SheetTitle>{t("pages.agent.cron.edit_title")}</SheetTitle>
        </SheetHeader>

        <div className="flex-1 space-y-5 overflow-y-auto px-4 py-2">
          <div className="space-y-1.5">
            <Label htmlFor="cron-edit-name">
              {t("pages.agent.cron.field_name")}
            </Label>
            <Input
              id="cron-edit-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="cron-edit-message">
              {job?.payload.deliver
                ? t("pages.agent.cron.field_message_static")
                : t("pages.agent.cron.field_message_dynamic")}
            </Label>
            <Textarea
              id="cron-edit-message"
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              rows={5}
              className="resize-none"
            />
            <p className="text-muted-foreground text-xs">
              {job?.payload.deliver
                ? t("pages.agent.cron.field_message_static_hint")
                : t("pages.agent.cron.field_message_dynamic_hint")}
            </p>
          </div>
        </div>

        <SheetFooter>
          <Button variant="outline" onClick={onClose} disabled={saving}>
            {t("common.cancel")}
          </Button>
          <Button
            onClick={() =>
              job &&
              onSave(job.id, { name: name.trim() || job.name, message })
            }
            disabled={saving || !job}
          >
            {saving && <IconLoader2 className="mr-1 size-4 animate-spin" />}
            {t("common.save")}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}

// ── Delete confirmation ───────────────────────────────────────────────────────

interface DeleteConfirmProps {
  job: CronJob | null
  onClose: () => void
  onConfirm: (id: string) => void
}

function DeleteConfirmDialog({ job, onClose, onConfirm }: DeleteConfirmProps) {
  const { t } = useTranslation()

  return (
    <AlertDialog open={!!job} onOpenChange={(open) => { if (!open) onClose() }}>
      <AlertDialogContent size="sm">
        <AlertDialogHeader>
          <AlertDialogTitle>
            {t("pages.agent.cron.delete_title")}
          </AlertDialogTitle>
          <AlertDialogDescription>
            {t("pages.agent.cron.delete_description").replace(
              "{{name}}",
              job?.name ?? "",
            )}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={onClose}>
            {t("common.cancel")}
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            onClick={() => job && onConfirm(job.id)}
          >
            {t("pages.agent.cron.delete_confirm")}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

// ── Main page ─────────────────────────────────────────────────────────────────

export function CronPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [editJob, setEditJob] = React.useState<CronJob | null>(null)
  const [deleteJob, setDeleteJob] = React.useState<CronJob | null>(null)

  const {
    data: jobs,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["cron-jobs"],
    queryFn: getCronJobs,
    refetchInterval: 5000,
  })

  const toggleMutation = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      updateCronJob(id, { enabled }),
    onSuccess: (_, { enabled }) => {
      toast.success(
        enabled
          ? t("pages.agent.cron.enable_success")
          : t("pages.agent.cron.disable_success"),
      )
      updateGatewayStore({ restartRequired: true })
      void queryClient.invalidateQueries({ queryKey: ["cron-jobs"] })
    },
    onError: (err) => {
      toast.error(
        err instanceof Error ? err.message : t("pages.agent.cron.toggle_error"),
      )
    },
  })

  const editMutation = useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: CronUpdateRequest }) =>
      updateCronJob(id, patch),
    onSuccess: () => {
      toast.success(t("pages.agent.cron.edit_success"))
      updateGatewayStore({ restartRequired: true })
      setEditJob(null)
      void queryClient.invalidateQueries({ queryKey: ["cron-jobs"] })
    },
    onError: (err) => {
      toast.error(
        err instanceof Error ? err.message : t("pages.agent.cron.edit_error"),
      )
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteCronJob(id),
    onSuccess: () => {
      toast.success(t("pages.agent.cron.delete_success"))
      updateGatewayStore({ restartRequired: true })
      setDeleteJob(null)
      void queryClient.invalidateQueries({ queryKey: ["cron-jobs"] })
    },
    onError: (err) => {
      toast.error(
        err instanceof Error ? err.message : t("pages.agent.cron.delete_error"),
      )
    },
  })

  return (
    <div className="flex h-full flex-col">
      <PageHeader title={t("navigation.cron")} />

      <div className="flex-1 overflow-auto px-6 py-3">
        <div className="w-full max-w-4xl space-y-4">
          {isLoading ? (
            <div className="text-muted-foreground flex items-center gap-2 py-6 text-sm">
              <IconLoader2 className="size-4 animate-spin" />
              {t("labels.loading")}
            </div>
          ) : error ? (
            <div className="text-destructive py-6 text-sm">
              {t("pages.agent.cron.load_error")}
            </div>
          ) : !jobs?.length ? (
            <Card className="border-dashed">
              <CardContent className="text-muted-foreground py-12 text-center text-sm">
                <IconClock className="mx-auto mb-3 size-8 opacity-30" />
                <p>{t("pages.agent.cron.empty")}</p>
                <p className="mt-1 text-xs opacity-70">
                  {t("pages.agent.cron.empty_hint")}
                </p>
              </CardContent>
            </Card>
          ) : (
            <>
              <p className="text-muted-foreground text-sm">
                {t("pages.agent.cron.description").replace(
                  "{{count}}",
                  String(jobs.length),
                )}
              </p>
              <div className="space-y-3">
                {jobs.map((job) => (
                  <CronJobCard
                    key={job.id}
                    job={job}
                    t={t}
                    onEdit={() => setEditJob(job)}
                    onDelete={() => setDeleteJob(job)}
                    onToggle={(enabled) =>
                      toggleMutation.mutate({ id: job.id, enabled })
                    }
                    toggling={
                      toggleMutation.isPending &&
                      toggleMutation.variables?.id === job.id
                    }
                  />
                ))}
              </div>
            </>
          )}
        </div>
      </div>

      <EditSheet
        job={editJob}
        onClose={() => setEditJob(null)}
        onSave={(id, patch) => editMutation.mutate({ id, patch })}
        saving={editMutation.isPending}
      />

      <DeleteConfirmDialog
        job={deleteJob}
        onClose={() => setDeleteJob(null)}
        onConfirm={(id) => deleteMutation.mutate(id)}
      />
    </div>
  )
}

// ── Job card ──────────────────────────────────────────────────────────────────

interface CronJobCardProps {
  job: CronJob
  t: (k: string) => string
  onEdit: () => void
  onDelete: () => void
  onToggle: (enabled: boolean) => void
  toggling: boolean
}

function CronJobCard({
  job,
  t,
  onEdit,
  onDelete,
  onToggle,
  toggling,
}: CronJobCardProps) {
  const scheduleLabel = formatSchedule(job, t)
  const nextRun = formatNextRun(job.state.nextRunAtMs, t)
  const lastRun = formatLastRun(job.state.lastRunAtMs)

  return (
    <Card
      className={cn(
        "gap-0 border transition-colors",
        job.enabled
          ? "border-border/60 bg-card"
          : "border-border/40 bg-muted/30 opacity-70",
      )}
    >
      <CardHeader className="pb-2">
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <CardTitle className="text-sm font-semibold">{job.name}</CardTitle>
              <StatusBadge enabled={job.enabled} t={t} />
              {job.state.lastStatus && (
                <LastStatusBadge status={job.state.lastStatus} t={t} />
              )}
              <ModeBadge deliver={job.payload.deliver} t={t} />
            </div>
            <CardDescription className="mt-1.5 line-clamp-2 break-words text-xs">
              {job.payload.message}
            </CardDescription>
          </div>

          <div className="flex shrink-0 items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0"
              onClick={onEdit}
              title={t("pages.agent.cron.edit")}
            >
              <IconPencil className="size-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0"
              onClick={() => onToggle(!job.enabled)}
              disabled={toggling}
              title={
                job.enabled
                  ? t("pages.agent.cron.disable")
                  : t("pages.agent.cron.enable")
              }
            >
              {toggling ? (
                <IconLoader2 className="size-3.5 animate-spin" />
              ) : job.enabled ? (
                <IconPlayerStop className="size-3.5" />
              ) : (
                <IconPlayerPlay className="size-3.5" />
              )}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="text-destructive hover:text-destructive h-8 w-8 p-0"
              onClick={onDelete}
              title={t("pages.agent.cron.delete")}
            >
              <IconTrash className="size-3.5" />
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent className="pb-3">
        <div className="text-muted-foreground grid grid-cols-2 gap-x-6 gap-y-1.5 text-xs sm:grid-cols-4">
          <div>
            <p className="font-medium">{t("pages.agent.cron.col_schedule")}</p>
            <p className="font-mono">{scheduleLabel}</p>
          </div>
          <div>
            <p className="font-medium">{t("pages.agent.cron.col_next_run")}</p>
            <p>{nextRun}</p>
          </div>
          <div>
            <p className="font-medium">{t("pages.agent.cron.col_last_run")}</p>
            <p>{lastRun}</p>
          </div>
          <div>
            <p className="font-medium">{t("pages.agent.cron.col_channel")}</p>
            <p className="truncate">{job.payload.channel || "—"}</p>
          </div>
        </div>
        {job.state.lastError && (
          <p className="text-destructive mt-2 truncate text-xs">
            {job.state.lastError}
          </p>
        )}
      </CardContent>
    </Card>
  )
}

function StatusBadge({
  enabled,
  t,
}: {
  enabled: boolean
  t: (k: string) => string
}) {
  return (
    <span
      className={cn(
        "rounded px-1.5 py-0.5 text-[10px] font-semibold",
        enabled
          ? "bg-emerald-100 text-emerald-700"
          : "bg-muted text-muted-foreground",
      )}
    >
      {enabled
        ? t("pages.agent.cron.status_enabled")
        : t("pages.agent.cron.status_disabled")}
    </span>
  )
}

function LastStatusBadge({
  status,
  t,
}: {
  status: string
  t: (k: string) => string
}) {
  const isError = status === "error"
  return (
    <span
      className={cn(
        "rounded px-1.5 py-0.5 text-[10px] font-semibold",
        isError ? "bg-red-100 text-red-700" : "bg-emerald-100 text-emerald-700",
      )}
    >
      {isError
        ? t("pages.agent.cron.last_status_error")
        : t("pages.agent.cron.last_status_ok")}
    </span>
  )
}

function ModeBadge({
  deliver,
  t,
}: {
  deliver: boolean
  t: (k: string) => string
}) {
  return (
    <span className="rounded bg-blue-50 px-1.5 py-0.5 text-[10px] font-semibold text-blue-600">
      {deliver
        ? t("pages.agent.cron.mode_static")
        : t("pages.agent.cron.mode_dynamic")}
    </span>
  )
}
