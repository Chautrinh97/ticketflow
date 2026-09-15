'use client'

import { CalendarX } from 'lucide-react'
import Link from 'next/link'
import { useState } from 'react'

import { OrderStatusBadge } from '@/components/data-display/Badge'
import { Tabs } from '@/components/data-display/Tabs'
import { EmptyState } from '@/components/feedback/EmptyState'
import { Skeleton } from '@/components/feedback/Skeleton'
import { Button } from '@/components/ui/Button'
import { useMyBookings } from '@/lib/hooks/useMyBookings'
import { formatCurrency } from '@/lib/utils/formatCurrency'
import { formatDateTime } from '@/lib/utils/formatDate'

const TABS = [
  { key: '', label: 'Tất cả' },
  { key: 'pending', label: 'Chờ thanh toán' },
  { key: 'paid', label: 'Đã thanh toán' },
  { key: 'cancelled', label: 'Đã huỷ' },
]

export default function MyBookingsPage() {
  const [status, setStatus] = useState('')
  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useMyBookings(status || undefined)
  const items = data?.pages.flatMap((p) => p.items) ?? []

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <h1 className="text-2xl font-bold text-gray-900">Vé của tôi</h1>
      <div className="mt-4">
        <Tabs tabs={TABS} activeKey={status} onChange={setStatus} />
      </div>
      <div className="mt-6">
        {isLoading ? (
          <div className="space-y-4">
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-20 w-full" />
            ))}
          </div>
        ) : items.length === 0 ? (
          <EmptyState icon={CalendarX} title="Chưa có đơn hàng nào" description="Vé bạn đặt sẽ hiển thị ở đây." />
        ) : (
          <div className="space-y-4">
            {items.map((order) => (
              <Link
                key={order.id}
                href={`/me/bookings/${order.id}`}
                className="flex items-center justify-between rounded-lg border border-gray-200 p-4 hover:bg-gray-50"
              >
                <div>
                  <p className="text-sm text-gray-500">{formatDateTime(order.created_at)}</p>
                  <p className="font-semibold text-gray-900">{formatCurrency(order.total_amount)}</p>
                </div>
                <OrderStatusBadge status={order.status} />
              </Link>
            ))}
            {hasNextPage && (
              <Button
                variant="secondary"
                className="w-full"
                loading={isFetchingNextPage}
                onClick={() => fetchNextPage()}
              >
                Xem thêm
              </Button>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
