import { IconTrash } from "@tabler/icons-react"
import { useMemo } from "react"
import { useTranslation } from "react-i18next"

import { LogsPanel } from "@/components/logs/logs-panel"
import { ToolLogsPanel } from "@/components/logs/tool-logs-panel"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useGatewayLogs } from "@/hooks/use-gateway-logs"
import { useLogWrapColumns } from "@/hooks/use-log-wrap-columns"
import { parseToolLogs } from "@/lib/tool-log-parser"

export function LogsPage() {
  const { t } = useTranslation()
  const { clearLogs, clearing, logs } = useGatewayLogs()
  const { contentRef, measureRef, wrapColumns } = useLogWrapColumns()

  const toolEntries = useMemo(() => parseToolLogs(logs), [logs])

  return (
    <div className="flex h-full flex-col">
      <PageHeader
        title={t("navigation.logs")}
        children={
          <Button
            variant="outline"
            size="sm"
            onClick={clearLogs}
            disabled={logs.length === 0 || clearing}
          >
            <IconTrash className="size-4" />
            {t("pages.logs.clear")}
          </Button>
        }
      />

      <Tabs defaultValue="all" className="flex flex-1 flex-col overflow-hidden">
        <TabsList>
          <TabsTrigger value="all">{t("pages.logs.tabs.all")}</TabsTrigger>
          <TabsTrigger value="tools">
            {t("pages.logs.tabs.tools")}
            {toolEntries.length > 0 && (
              <span className="ml-1 text-xs opacity-60">({toolEntries.length})</span>
            )}
          </TabsTrigger>
        </TabsList>
        <TabsContent value="all" className="flex flex-1 flex-col overflow-hidden p-4 sm:p-8">
          <LogsPanel
            logs={logs}
            wrapColumns={wrapColumns}
            contentRef={contentRef}
            measureRef={measureRef}
          />
        </TabsContent>
        <TabsContent value="tools" className="flex flex-1 flex-col overflow-hidden p-4 sm:p-8">
          <ToolLogsPanel entries={toolEntries} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
