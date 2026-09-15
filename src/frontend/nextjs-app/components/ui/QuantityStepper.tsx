'use client'

import { Minus, Plus } from 'lucide-react'

import { IconButton } from './IconButton'

interface QuantityStepperProps {
  value: number
  min?: number
  max: number
  onChange: (value: number) => void
}

// docs/06-frontend/components/buttons-inputs.md: min/max clamp on blur,
// minus disabled at min, plus disabled at max. Default min=0 here (not the
// documented min=1) because Phase 1's checkout lets each ticket-type row
// be independently included/excluded in one multi-item order — 0 means
// "not selected", see components/data-display/TicketTypeRow.tsx.
export function QuantityStepper({ value, min = 0, max, onChange }: QuantityStepperProps) {
  const clamp = (v: number) => Math.min(max, Math.max(min, v))

  return (
    <div className="flex items-center gap-3">
      <IconButton
        variant="ghost"
        size="sm"
        aria-label="Giảm số lượng"
        disabled={value <= min}
        onClick={() => onChange(clamp(value - 1))}
      >
        <Minus className="h-4 w-4" />
      </IconButton>
      <input
        type="number"
        className="w-12 text-center text-lg font-semibold focus:outline-none"
        value={value}
        onChange={(e) => {
          const n = Number(e.target.value)
          if (!Number.isNaN(n)) onChange(n)
        }}
        onBlur={(e) => onChange(clamp(Number(e.target.value) || min))}
      />
      <IconButton
        variant="ghost"
        size="sm"
        aria-label="Tăng số lượng"
        disabled={value >= max}
        onClick={() => onChange(clamp(value + 1))}
      >
        <Plus className="h-4 w-4" />
      </IconButton>
    </div>
  )
}
