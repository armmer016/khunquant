import {
  IconFileText,
  IconFolder,
  IconLoader2,
  IconPlus,
  IconTrash,
} from "@tabler/icons-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  type AgentMemoryFile,
  createMemoryFile,
  deleteMemoryFile,
  getMemoryFile,
  getMemoryFiles,
  getMemorySize,
  saveMemoryFile,
} from "@/api/agent-memory"
import { PageHeader } from "@/components/page-header"
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"

import { DCASnapshotPanel } from "./dca-snapshot-panel"
import { SnapshotPanel } from "./snapshot-panel"

function formatBytes(bytes: number): string {
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
}

interface GroupedFiles {
  root: AgentMemoryFile[]
  months: { label: string; files: AgentMemoryFile[] }[]
}

function groupFiles(files: AgentMemoryFile[]): GroupedFiles {
  const root: AgentMemoryFile[] = []
  const monthMap = new Map<string, AgentMemoryFile[]>()

  for (const f of files) {
    const parts = f.path.split("/")
    if (parts.length === 1) {
      root.push(f)
    } else {
      const month = parts[0]
      if (!monthMap.has(month)) monthMap.set(month, [])
      monthMap.get(month)!.push(f)
    }
  }

  // Sort month keys descending (newest first)
  const months = Array.from(monthMap.entries())
    .sort(([a], [b]) => b.localeCompare(a))
    .map(([label, files]) => ({
      label,
      files: files.sort((a, b) => b.name.localeCompare(a.name)),
    }))

  return { root, months }
}

export function AgentMemoryPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const [selectedPath, setSelectedPath] = useState<string | null>(null)
  const [editorValue, setEditorValue] = useState("")
  const [isDirty, setIsDirty] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [showNewFileDialog, setShowNewFileDialog] = useState(false)
  const [newFilePath, setNewFilePath] = useState("")

  const { data: sizeInfo } = useQuery({
    queryKey: ["agent-memory-size"],
    queryFn: getMemorySize,
  })

  const { data: files, isLoading: isFilesLoading } = useQuery({
    queryKey: ["agent-memory-files"],
    queryFn: getMemoryFiles,
  })

  const { data: fileContent, isLoading: isFileLoading } = useQuery({
    queryKey: ["agent-memory-file", selectedPath],
    queryFn: () => getMemoryFile(selectedPath!),
    enabled: selectedPath !== null,
  })

  const grouped = useMemo(() => groupFiles(files ?? []), [files])

  const effectiveEditorValue = isDirty ? editorValue : (fileContent?.content ?? "")

  const handleSelectFile = (path: string) => {
    if (isDirty) {
      if (!window.confirm(t("pages.agent.agent_memory.unsaved_confirm"))) return
    }
    setSelectedPath(path)
    setEditorValue("")
    setIsDirty(false)
  }

  const saveMutation = useMutation({
    mutationFn: () => saveMemoryFile(selectedPath!, effectiveEditorValue),
    onSuccess: () => {
      toast.success(t("pages.agent.agent_memory.save_success"))
      setIsDirty(false)
      void queryClient.invalidateQueries({ queryKey: ["agent-memory-files"] })
      void queryClient.invalidateQueries({ queryKey: ["agent-memory-file", selectedPath] })
      void queryClient.invalidateQueries({ queryKey: ["agent-memory-size"] })
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : t("pages.agent.agent_memory.save_error"))
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteMemoryFile(selectedPath!),
    onSuccess: () => {
      toast.success(t("pages.agent.agent_memory.delete_success"))
      setShowDeleteDialog(false)
      setSelectedPath(null)
      setEditorValue("")
      setIsDirty(false)
      void queryClient.invalidateQueries({ queryKey: ["agent-memory-files"] })
      void queryClient.invalidateQueries({ queryKey: ["agent-memory-size"] })
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : t("pages.agent.agent_memory.delete_error"))
    },
  })

  const createMutation = useMutation({
    mutationFn: () => {
      const path = newFilePath.endsWith(".md") ? newFilePath : `${newFilePath}.md`
      return createMemoryFile(path, "")
    },
    onSuccess: (result) => {
      toast.success(t("pages.agent.agent_memory.create_success"))
      setShowNewFileDialog(false)
      setNewFilePath("")
      void queryClient.invalidateQueries({ queryKey: ["agent-memory-files"] })
      void queryClient.invalidateQueries({ queryKey: ["agent-memory-size"] })
      setSelectedPath(result.path)
      setEditorValue("")
      setIsDirty(false)
    },
    onError: (err) => {
      toast.error(err instanceof Error ? err.message : t("pages.agent.agent_memory.create_error"))
    },
  })

  const fileItemClass = (path: string) =>
    `flex w-full items-center gap-2 rounded-md px-3 py-1.5 text-left text-sm transition-colors ${
      selectedPath === path
        ? "bg-accent/80 text-foreground font-medium"
        : "text-muted-foreground hover:bg-muted/60"
    }`

  return (
    <div className="flex h-full flex-col">
      <PageHeader
        title={t("navigation.agent_memory")}
        titleExtra={
          sizeInfo != null ? (
            <span className="text-muted-foreground text-sm font-normal">
              (~{formatBytes(sizeInfo.total_bytes)})
            </span>
          ) : undefined
        }
      />

      <Tabs defaultValue="general" className="flex flex-1 flex-col overflow-hidden px-4 pb-4">
        <TabsList className="mb-2 shrink-0 self-start">
          <TabsTrigger value="general">
            {t("pages.agent.agent_memory.tabs.general")}
            {sizeInfo != null && (
              <span className="text-muted-foreground ml-1.5 text-xs font-normal">
                ({formatBytes(sizeInfo.general_bytes)})
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="snapshot">
            {t("pages.agent.agent_memory.tabs.snapshot")}
            {sizeInfo != null && (
              <span className="text-muted-foreground ml-1.5 text-xs font-normal">
                ({formatBytes(sizeInfo.snapshot_bytes)})
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="dca_snapshot">
            {t("pages.agent.agent_memory.tabs.dca_snapshot")}
            {sizeInfo != null && (
              <span className="text-muted-foreground ml-1.5 text-xs font-normal">
                ({formatBytes(sizeInfo.dca_bytes)})
              </span>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-0 flex flex-1 flex-col overflow-hidden">
          {/* General tab header with New File button */}
          <div className="mb-2 flex shrink-0 justify-end">
            <Button variant="outline" size="sm" onClick={() => setShowNewFileDialog(true)}>
              <IconPlus className="size-4" />
              {t("pages.agent.agent_memory.new_file")}
            </Button>
          </div>

          <div className="border-border/40 flex min-h-0 flex-1 overflow-hidden rounded-lg border">
            {/* Left panel */}
            <div className="border-border/40 flex w-56 shrink-0 flex-col border-r">
              <div className="flex-1 overflow-auto p-2">
                {isFilesLoading ? (
                  <div className="text-muted-foreground p-2 text-sm">{t("labels.loading")}</div>
                ) : (
                  <>
                    {/* Root files */}
                    {grouped.root.length > 0 && (
                      <ul className="mb-2 space-y-0.5">
                        {grouped.root.map((f) => (
                          <li key={f.path}>
                            <button onClick={() => handleSelectFile(f.path)} className={fileItemClass(f.path)}>
                              <IconFileText className="size-3.5 shrink-0 opacity-60" />
                              <span className="truncate">{f.name}</span>
                            </button>
                          </li>
                        ))}
                      </ul>
                    )}

                    {/* Monthly groups */}
                    {grouped.months.map((group) => (
                      <div key={group.label} className="mb-2">
                        <div className="text-muted-foreground/60 mb-0.5 flex items-center gap-1.5 px-3 py-1 text-[11px] font-semibold uppercase tracking-wider">
                          <IconFolder className="size-3 shrink-0" />
                          {group.label}
                        </div>
                        <ul className="space-y-0.5">
                          {group.files.map((f) => (
                            <li key={f.path}>
                              <button onClick={() => handleSelectFile(f.path)} className={fileItemClass(f.path)}>
                                <IconFileText className="size-3.5 shrink-0 opacity-60" />
                                <span className="truncate">{f.name}</span>
                              </button>
                            </li>
                          ))}
                        </ul>
                      </div>
                    ))}

                    {grouped.root.length === 0 && grouped.months.length === 0 && (
                      <div className="text-muted-foreground p-2 text-sm">
                        {t("pages.agent.agent_memory.empty")}
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>

            {/* Right panel: editor */}
            <div className="flex min-h-0 flex-1 flex-col p-4">
              {selectedPath ? (
                <div className="flex min-h-0 flex-1 flex-col gap-3">
                  <div className="flex shrink-0 items-center justify-between">
                    <h3 className="text-foreground/90 font-mono text-sm font-medium">
                      {selectedPath}
                    </h3>
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      className="text-muted-foreground hover:text-destructive"
                      onClick={() => setShowDeleteDialog(true)}
                      title={t("pages.agent.agent_memory.delete")}
                    >
                      <IconTrash className="size-4" />
                    </Button>
                  </div>

                  {isDirty && (
                    <div className="shrink-0 rounded-lg border border-yellow-200 bg-yellow-50 p-2 text-sm text-yellow-700">
                      {t("pages.agent.agent_memory.unsaved_changes")}
                    </div>
                  )}

                  <div className="relative min-h-0 flex-1 overflow-hidden rounded-lg border shadow-sm">
                    {isFileLoading ? (
                      <div className="flex h-full items-center justify-center">
                        <IconLoader2 className="text-muted-foreground size-5 animate-spin" />
                      </div>
                    ) : (
                      <Textarea
                        value={effectiveEditorValue}
                        onChange={(e) => {
                          setEditorValue(e.target.value)
                          setIsDirty(true)
                        }}
                        wrap="off"
                        className="h-full min-h-0 resize-none overflow-auto border-0 bg-transparent px-4 py-3 font-mono text-sm [overflow-wrap:normal] whitespace-pre shadow-none focus-visible:ring-0"
                        placeholder={t("pages.agent.agent_memory.placeholder")}
                      />
                    )}
                  </div>

                  <div className="flex shrink-0 justify-end">
                    <Button
                      onClick={() => saveMutation.mutate()}
                      disabled={saveMutation.isPending || !isDirty}
                    >
                      {saveMutation.isPending ? t("common.saving") : t("common.save")}
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
                  {t("pages.agent.agent_memory.select_hint")}
                </div>
              )}
            </div>
          </div>
        </TabsContent>

        <TabsContent value="snapshot" className="mt-0 flex flex-1 overflow-hidden">
          <div className="border-border/40 flex min-h-0 flex-1 overflow-hidden rounded-lg border">
            <SnapshotPanel
              onDeleteSuccess={() =>
                void queryClient.invalidateQueries({ queryKey: ["agent-memory-size"] })
              }
            />
          </div>
        </TabsContent>

        <TabsContent value="dca_snapshot" className="mt-0 flex flex-1 overflow-hidden">
          <div className="border-border/40 flex min-h-0 flex-1 overflow-hidden rounded-lg border">
            <DCASnapshotPanel />
          </div>
        </TabsContent>
      </Tabs>

      {/* Delete file dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogTitle>{t("pages.agent.agent_memory.delete_title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("pages.agent.agent_memory.delete_description", { path: selectedPath })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMutation.isPending}>{t("common.cancel")}</AlertDialogCancel>
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
              {t("pages.agent.agent_memory.delete_confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* New file dialog */}
      <AlertDialog open={showNewFileDialog} onOpenChange={setShowNewFileDialog}>
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogTitle>{t("pages.agent.agent_memory.new_file_title")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("pages.agent.agent_memory.new_file_description")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="px-1 py-2">
            <input
              type="text"
              value={newFilePath}
              onChange={(e) => setNewFilePath(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && newFilePath.trim()) createMutation.mutate()
              }}
              placeholder="NOTE.md or 202603/note.md"
              className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex h-10 w-full rounded-md border px-3 py-2 font-mono text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              autoFocus
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={createMutation.isPending} onClick={() => setNewFilePath("")}>
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={createMutation.isPending || !newFilePath.trim()}
              onClick={() => createMutation.mutate()}
            >
              {createMutation.isPending ? <IconLoader2 className="size-4 animate-spin" /> : null}
              {t("pages.agent.agent_memory.create_confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
