import {
  IconFileText,
  IconLoader2,
  IconPlus,
  IconTrash,
} from "@tabler/icons-react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  createAgentConfigFile,
  deleteAgentConfigFile,
  getAgentConfigFile,
  getAgentConfigFiles,
  saveAgentConfigFile,
} from "@/api/agent-config"
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
import { Textarea } from "@/components/ui/textarea"

export function AgentConfigPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [editorValue, setEditorValue] = useState("")
  const [isDirty, setIsDirty] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const [showNewFileDialog, setShowNewFileDialog] = useState(false)
  const [newFileName, setNewFileName] = useState("")

  const { data: files, isLoading: isFilesLoading } = useQuery({
    queryKey: ["agent-config-files"],
    queryFn: getAgentConfigFiles,
  })

  const { data: fileContent, isLoading: isFileLoading } = useQuery({
    queryKey: ["agent-config-file", selectedFile],
    queryFn: () => getAgentConfigFile(selectedFile!),
    enabled: selectedFile !== null,
  })

  const effectiveEditorValue =
    isDirty ? editorValue : (fileContent?.content ?? "")

  const handleSelectFile = (name: string) => {
    if (isDirty) {
      const confirmed = window.confirm(
        t("pages.agent.agent_config.unsaved_confirm"),
      )
      if (!confirmed) return
    }
    setSelectedFile(name)
    setEditorValue("")
    setIsDirty(false)
  }

  const saveMutation = useMutation({
    mutationFn: () =>
      saveAgentConfigFile(selectedFile!, effectiveEditorValue),
    onSuccess: () => {
      toast.success(t("pages.agent.agent_config.save_success"))
      setIsDirty(false)
      void queryClient.invalidateQueries({ queryKey: ["agent-config-files"] })
      void queryClient.invalidateQueries({
        queryKey: ["agent-config-file", selectedFile],
      })
    },
    onError: (err) => {
      toast.error(
        err instanceof Error
          ? err.message
          : t("pages.agent.agent_config.save_error"),
      )
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteAgentConfigFile(selectedFile!),
    onSuccess: () => {
      toast.success(t("pages.agent.agent_config.delete_success"))
      setShowDeleteDialog(false)
      setSelectedFile(null)
      setEditorValue("")
      setIsDirty(false)
      void queryClient.invalidateQueries({ queryKey: ["agent-config-files"] })
    },
    onError: (err) => {
      toast.error(
        err instanceof Error
          ? err.message
          : t("pages.agent.agent_config.delete_error"),
      )
    },
  })

  const createMutation = useMutation({
    mutationFn: () => {
      const name = newFileName.endsWith(".md")
        ? newFileName
        : `${newFileName}.md`
      return createAgentConfigFile(name, "")
    },
    onSuccess: (result) => {
      toast.success(t("pages.agent.agent_config.create_success"))
      setShowNewFileDialog(false)
      setNewFileName("")
      void queryClient.invalidateQueries({ queryKey: ["agent-config-files"] })
      setSelectedFile(result.name)
      setEditorValue("")
      setIsDirty(false)
    },
    onError: (err) => {
      toast.error(
        err instanceof Error
          ? err.message
          : t("pages.agent.agent_config.create_error"),
      )
    },
  })

  return (
    <div className="flex h-full flex-col">
      <PageHeader title={t("navigation.agent_config")}>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setShowNewFileDialog(true)}
        >
          <IconPlus className="size-4" />
          {t("pages.agent.agent_config.new_file")}
        </Button>
      </PageHeader>

      <div className="flex min-h-0 flex-1 overflow-hidden">
        {/* Left panel: file list */}
        <div className="border-border/40 flex w-56 shrink-0 flex-col border-r">
          <div className="flex-1 overflow-auto p-2">
            {isFilesLoading ? (
              <div className="text-muted-foreground p-2 text-sm">
                {t("labels.loading")}
              </div>
            ) : files && files.length > 0 ? (
              <ul className="space-y-0.5">
                {files.map((file) => (
                  <li key={file.name}>
                    <button
                      onClick={() => handleSelectFile(file.name)}
                      className={`flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors ${
                        selectedFile === file.name
                          ? "bg-accent/80 text-foreground font-medium"
                          : "text-muted-foreground hover:bg-muted/60"
                      }`}
                    >
                      <IconFileText className="size-3.5 shrink-0 opacity-60" />
                      <span className="truncate">{file.name}</span>
                    </button>
                  </li>
                ))}
              </ul>
            ) : (
              <div className="text-muted-foreground p-2 text-sm">
                {t("pages.agent.agent_config.empty")}
              </div>
            )}
          </div>
        </div>

        {/* Right panel: editor */}
        <div className="flex min-h-0 flex-1 flex-col p-4">
          {selectedFile ? (
            <div className="flex min-h-0 flex-1 flex-col gap-3">
              <div className="flex shrink-0 items-center justify-between">
                <h3 className="text-foreground/90 font-medium">
                  {selectedFile}
                </h3>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  className="text-muted-foreground hover:text-destructive"
                  onClick={() => setShowDeleteDialog(true)}
                  title={t("pages.agent.agent_config.delete")}
                >
                  <IconTrash className="size-4" />
                </Button>
              </div>

              {isDirty && (
                <div className="shrink-0 rounded-lg border border-yellow-200 bg-yellow-50 p-2 text-sm text-yellow-700">
                  {t("pages.agent.agent_config.unsaved_changes")}
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
                    placeholder={t("pages.agent.agent_config.placeholder")}
                  />
                )}
              </div>

              <div className="flex shrink-0 justify-end">
                <Button
                  onClick={() => saveMutation.mutate()}
                  disabled={saveMutation.isPending || !isDirty}
                >
                  {saveMutation.isPending
                    ? t("common.saving")
                    : t("common.save")}
                </Button>
              </div>
            </div>
          ) : (
            <div className="text-muted-foreground flex h-full items-center justify-center text-sm">
              {t("pages.agent.agent_config.select_hint")}
            </div>
          )}
        </div>
      </div>

      {/* Delete dialog */}
      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("pages.agent.agent_config.delete_title")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("pages.agent.agent_config.delete_description", {
                name: selectedFile,
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
              {t("pages.agent.agent_config.delete_confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* New file dialog */}
      <AlertDialog open={showNewFileDialog} onOpenChange={setShowNewFileDialog}>
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("pages.agent.agent_config.new_file_title")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("pages.agent.agent_config.new_file_description")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="px-1 py-2">
            <input
              type="text"
              value={newFileName}
              onChange={(e) => setNewFileName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && newFileName.trim()) {
                  createMutation.mutate()
                }
              }}
              placeholder="CUSTOM.md"
              className="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex h-10 w-full rounded-md border px-3 py-2 text-sm file:border-0 file:bg-transparent file:text-sm file:font-medium focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              autoFocus
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel
              disabled={createMutation.isPending}
              onClick={() => setNewFileName("")}
            >
              {t("common.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction
              disabled={createMutation.isPending || !newFileName.trim()}
              onClick={() => createMutation.mutate()}
            >
              {createMutation.isPending ? (
                <IconLoader2 className="size-4 animate-spin" />
              ) : null}
              {t("pages.agent.agent_config.create_confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
