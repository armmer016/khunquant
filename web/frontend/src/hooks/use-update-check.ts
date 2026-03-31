import { useCallback, useEffect, useRef, useState } from "react"

import { type UpdateStatus, getUpdateStatus } from "@/api/update"

const QUICK_RETRY_INTERVAL_MS = 10 * 1000 // 10 seconds
const QUICK_RETRY_COUNT = 6               // up to ~60 seconds of quick retries
const POLL_INTERVAL_MS = 5 * 60 * 1000   // 5 minutes

interface UseUpdateCheckResult {
  status: UpdateStatus | null
  refetch: () => void
}

export function useUpdateCheck(): UseUpdateCheckResult {
  const [status, setStatus] = useState<UpdateStatus | null>(null)
  const quickAttemptsRef = useRef(0)
  const slowIdRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const doFetch = useCallback(() => {
    getUpdateStatus()
      .then((data) => {
        setStatus(data.is_outdated ? data : null)
      })
      .catch(() => {
        // silently ignore network errors
      })
  }, [])

  useEffect(() => {
    doFetch()

    const quickId = setInterval(() => {
      quickAttemptsRef.current++
      doFetch()
      if (quickAttemptsRef.current >= QUICK_RETRY_COUNT) {
        clearInterval(quickId)
        slowIdRef.current = setInterval(doFetch, POLL_INTERVAL_MS)
      }
    }, QUICK_RETRY_INTERVAL_MS)

    return () => {
      clearInterval(quickId)
      if (slowIdRef.current !== null) clearInterval(slowIdRef.current)
    }
  }, [doFetch])

  return { status, refetch: doFetch }
}
