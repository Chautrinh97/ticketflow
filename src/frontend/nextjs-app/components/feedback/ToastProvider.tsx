'use client'

import { CircleCheck, CircleX, X } from 'lucide-react'
import { createContext, useCallback, useContext, useRef, useState } from 'react'

type ToastVariant = 'success' | 'danger'

interface ToastItem {
  id: number
  variant: ToastVariant
  message: string
}

interface ToastContextValue {
  showToast: (variant: ToastVariant, message: string) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

// Per docs/06-frontend/interaction-patterns.md: success auto-dismiss 4s,
// danger 6s, max 3 concurrent (oldest drops first).
const MAX_TOASTS = 3
const DURATIONS: Record<ToastVariant, number> = { success: 4000, danger: 6000 }

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([])
  const idRef = useRef(0)

  const removeToast = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const showToast = useCallback(
    (variant: ToastVariant, message: string) => {
      const id = ++idRef.current
      setToasts((prev) => [{ id, variant, message }, ...prev].slice(0, MAX_TOASTS))
      setTimeout(() => removeToast(id), DURATIONS[variant])
    },
    [removeToast]
  )

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      <div className="fixed right-4 top-4 z-50 flex w-full max-w-sm flex-col gap-2">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={`flex items-start gap-3 rounded-lg border-l-4 bg-white p-4 shadow-md ${
              t.variant === 'success' ? 'border-green-600' : 'border-red-600'
            }`}
          >
            {t.variant === 'success' ? (
              <CircleCheck className="h-5 w-5 shrink-0 text-green-600" />
            ) : (
              <CircleX className="h-5 w-5 shrink-0 text-red-600" />
            )}
            <p className="flex-1 text-sm text-gray-900">{t.message}</p>
            <button aria-label="Đóng" onClick={() => removeToast(t.id)} className="text-gray-400 hover:text-gray-600">
              <X className="h-4 w-4" />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx
}
