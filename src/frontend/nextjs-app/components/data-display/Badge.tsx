import { cn } from '@/lib/utils/cn'

type Tone = 'success' | 'warning' | 'danger' | 'neutral'

const toneClasses: Record<Tone, string> = {
  success: 'bg-green-50 text-green-700',
  warning: 'bg-amber-50 text-amber-700',
  danger: 'bg-red-50 text-red-700',
  neutral: 'bg-gray-100 text-gray-600',
}

export function Badge({ tone, children }: { tone: Tone; children: React.ReactNode }) {
  return (
    <span className={cn('inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium', toneClasses[tone])}>
      {children}
    </span>
  )
}

// Status -> token/label maps mirror docs/06-frontend/components/data-display.md's
// table (enum source of truth: docs/03-data/postgres-schema.md).
const eventStatusTone: Record<string, Tone> = { draft: 'neutral', published: 'success', cancelled: 'danger' }
const orderStatusTone: Record<string, Tone> = { pending: 'warning', paid: 'success', cancelled: 'danger', expired: 'danger' }
const ticketStatusTone: Record<string, Tone> = { valid: 'success', used: 'neutral', cancelled: 'danger' }

const eventStatusLabel: Record<string, string> = { draft: 'Nháp', published: 'Đã xuất bản', cancelled: 'Đã huỷ' }
const orderStatusLabel: Record<string, string> = {
  pending: 'Chờ thanh toán',
  paid: 'Đã thanh toán',
  cancelled: 'Đã huỷ',
  expired: 'Hết hạn',
}
const ticketStatusLabel: Record<string, string> = { valid: 'Hợp lệ', used: 'Đã dùng', cancelled: 'Đã huỷ' }

export function EventStatusBadge({ status }: { status: string }) {
  return <Badge tone={eventStatusTone[status] ?? 'neutral'}>{eventStatusLabel[status] ?? status}</Badge>
}

export function OrderStatusBadge({ status }: { status: string }) {
  return <Badge tone={orderStatusTone[status] ?? 'neutral'}>{orderStatusLabel[status] ?? status}</Badge>
}

export function TicketStatusBadge({ status }: { status: string }) {
  return <Badge tone={ticketStatusTone[status] ?? 'neutral'}>{ticketStatusLabel[status] ?? status}</Badge>
}
