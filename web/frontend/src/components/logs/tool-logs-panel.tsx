import { IconAlertTriangle, IconCheck, IconTool } from "@tabler/icons-react"
import { useEffect, useRef, useState } from "react"

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import { ScrollArea } from "@/components/ui/scroll-area"
import type { ToolLogEntry } from "@/lib/tool-log-parser"

type ToolLogsPanelProps = {
  entries: ToolLogEntry[]
}

export function ToolLogsPanel({ entries }: ToolLogsPanelProps) {
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollIntoView({ behavior: "smooth" })
    }
  }, [entries])

  return (
    <div className="relative flex-1 overflow-hidden rounded-lg border border-zinc-800 bg-zinc-950 text-zinc-100">
      <ScrollArea className="h-full">
        <div className="p-4 space-y-2">
          {entries.length === 0 ? (
            <div className="text-zinc-500 italic">No tool calls yet.</div>
          ) : (
            entries.map((entry, index) => (
              <ToolLogCard key={index} entry={entry} />
            ))
          )}
          <div ref={scrollRef} />
        </div>
      </ScrollArea>
    </div>
  )
}

function ToolLogCard({ entry }: { entry: ToolLogEntry }) {
  const [open, setOpen] = useState(false)

  if (entry.kind === "call") {
    return (
      <Collapsible open={open} onOpenChange={setOpen}>
        <CollapsibleTrigger className="w-full text-left">
          <div className="flex items-center gap-2 rounded px-3 py-2 bg-zinc-900 hover:bg-zinc-800 transition-colors cursor-pointer">
            <IconTool className="size-4 text-zinc-400 shrink-0" />
            <span className="font-mono text-sm font-medium text-zinc-100">{entry.tool}</span>
            {entry.time && (
              <span className="ml-auto text-xs text-zinc-500">{entry.time}</span>
            )}
          </div>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <div className="px-3 py-2 bg-zinc-900/50 border-t border-zinc-800 rounded-b">
            <pre className="text-xs text-zinc-300 whitespace-pre-wrap break-all font-mono">
              {formatArgs(entry.argsPreview)}
            </pre>
          </div>
        </CollapsibleContent>
      </Collapsible>
    )
  }

  // result entry
  const Icon = entry.hasError ? IconAlertTriangle : IconCheck
  const iconColor = entry.hasError ? "text-red-400" : "text-green-400"
  const preview = entry.resultPreview + (entry.resultPreview.endsWith("...") ? " (truncated)" : "")

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className="w-full text-left">
        <div className="flex items-center gap-2 rounded px-3 py-2 bg-zinc-900 hover:bg-zinc-800 transition-colors cursor-pointer">
          <Icon className={`size-4 shrink-0 ${iconColor}`} />
          <span className="font-mono text-sm font-medium text-zinc-100">{entry.tool}</span>
          <span className="text-xs text-zinc-500">result</span>
          {entry.time && (
            <span className="ml-auto text-xs text-zinc-500">{entry.time}</span>
          )}
        </div>
      </CollapsibleTrigger>
      <CollapsibleContent>
        <div className="px-3 py-2 bg-zinc-900/50 border-t border-zinc-800 rounded-b">
          <pre className="text-xs text-zinc-300 whitespace-pre-wrap break-all font-mono">
            {preview || "(empty)"}
          </pre>
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

function formatArgs(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}
