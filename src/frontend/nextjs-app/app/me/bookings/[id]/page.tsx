'use client'

import { useQuery } from '@tanstack/react-query'
import { QRCodeSVG } from 'qrcode.react'

import { OrderStatusBadge, TicketStatusBadge } from '@/components/data-display/Badge'
import { Skeleton } from '@/components/feedback/Skeleton'
import { getBooking } from '@/lib/api/bookings'
import { formatCurrency } from '@/lib/utils/formatCurrency'
import { formatDateTime } from '@/lib/utils/formatDate'

// View-only in Phase 1 — no cancel action (POST /bookings/:id/cancel is
// Phase 2 per docs/02-domains/booking/spec.md's phase table).
export default function BookingDetailPage({ params }: { params: { id: string } }) {
  const { data: order, isLoading } = useQuery({
    queryKey: ['booking', params.id],
    queryFn: () => getBooking(params.id),
  })

  if (isLoading || !order) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-8">
        <Skeleton className="h-64 w-full" />
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-gray-900">Đơn hàng #{order.id.slice(0, 8)}</h1>
        <OrderStatusBadge status={order.status} />
      </div>
      <p className="mt-1 text-sm text-gray-500">{formatDateTime(order.created_at)}</p>
      <p className="mt-4 font-semibold text-gray-900">Tổng cộng: {formatCurrency(order.total_amount)}</p>

      {order.tickets.length > 0 && (
        <div className="mt-8">
          <h2 className="text-lg font-semibold text-gray-900">Vé của bạn</h2>
          <div className="mt-4 grid gap-4 sm:grid-cols-2">
            {order.tickets.map((ticket) => (
              <div key={ticket.id} className="flex flex-col items-center gap-3 rounded-lg border border-gray-200 p-4">
                <QRCodeSVG value={ticket.ticket_code} size={140} />
                <p className="font-mono text-sm text-gray-700">{ticket.ticket_code}</p>
                <TicketStatusBadge status={ticket.status} />
              </div>
            ))}
          </div>
        </div>
      )}

      {order.status === 'pending' && (
        <p className="mt-8 text-sm text-amber-600">
          Đơn hàng đang chờ thanh toán, hết hạn lúc {formatDateTime(order.expires_at)}.
        </p>
      )}
    </div>
  )
}
