import { cn } from "@/lib/utils"

interface AlertDotProps {
  className?: string
}

export function AlertDot({ className }: AlertDotProps) {
  return (
    <span className={cn("relative flex size-2 shrink-0", className)}>
      <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-red-400 opacity-75" />
      <span className="relative inline-flex size-2 rounded-full bg-red-500" />
    </span>
  )
}
