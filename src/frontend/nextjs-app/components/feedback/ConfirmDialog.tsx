'use client'

import { TriangleAlert } from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/Button'

interface ConfirmDialogProps {
  open: boolean
  title: string
  description: string
  confirmLabel?: string
  onConfirm: () => Promise<void> | void
  onCancel: () => void
}

// Per docs/06-frontend/interaction-patterns.md: mandatory before destructive
// actions, danger button always on the right, closes + shows a Toast after
// the API responds (the caller's onConfirm is responsible for that toast).
export function ConfirmDialog({ open, title, description, confirmLabel = 'Xác nhận', onConfirm, onCancel }: ConfirmDialogProps) {
  const [loading, setLoading] = useState(false)

  if (!open) return null

  async function handleConfirm() {
    setLoading(true)
    try {
      await onConfirm()
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={onCancel}>
      <div className="w-full max-w-sm rounded-xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
        <TriangleAlert className="h-6 w-6 text-red-600" />
        <h3 className="mt-3 text-lg font-semibold text-gray-900">{title}</h3>
        <p className="mt-2 text-sm text-gray-500">{description}</p>
        <div className="mt-6 flex justify-end gap-3">
          <Button variant="secondary" onClick={onCancel} disabled={loading}>
            Huỷ
          </Button>
          <Button variant="danger" onClick={handleConfirm} loading={loading}>
            {confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  )
}
