import { QuantityStepper } from '@/components/ui/QuantityStepper'
import { formatCurrency } from '@/lib/utils/formatCurrency'
import type { TicketType } from '@/types/api'

interface TicketTypeRowProps {
  ticketType: TicketType
  quantity?: number
  onQuantityChange?: (value: number) => void
  soldOverride?: string // event-manage.md's "sold_count/quota" display
}

export function TicketTypeRow({ ticketType, quantity, onQuantityChange, soldOverride }: TicketTypeRowProps) {
  const remaining = ticketType.quota - ticketType.sold_count
  const soldOut = remaining <= 0
  const low = !soldOut && remaining <= ticketType.quota * 0.1

  return (
    <div className="flex items-center justify-between border-b border-gray-200 py-4 last:border-b-0">
      <div>
        <h3 className="text-lg font-semibold text-gray-900">{ticketType.name}</h3>
        {soldOverride ? (
          <p className="text-sm text-gray-500">{soldOverride}</p>
        ) : (
          <p className={`text-sm ${soldOut ? 'text-red-600' : low ? 'text-amber-600' : 'text-gray-500'}`}>
            {soldOut ? 'Hết vé' : `Còn ${remaining} vé`}
          </p>
        )}
      </div>
      <div className="flex items-center gap-4">
        <span className="font-semibold text-gray-900">{formatCurrency(ticketType.price, ticketType.currency)}</span>
        {onQuantityChange && (
          <QuantityStepper value={quantity ?? 0} min={0} max={Math.min(10, remaining)} onChange={onQuantityChange} />
        )}
      </div>
    </div>
  )
}
