'use client'

import { useQuery } from '@tanstack/react-query'
import { CircleCheck, CircleX, Clock } from 'lucide-react'
import Link from 'next/link'
import { useEffect, useRef, useState } from 'react'

import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { getBooking } from '@/lib/api/bookings'
import { formatCurrency } from '@/lib/utils/formatCurrency'

const POLL_INTERVAL_MS = 3000
const POLL_TIMEOUT_MS = 2 * 60 * 1000

// Polls GET /bookings/:id every 3s while status=pending, capped at 2 min,
// per the implementation plan's frontend section. Under Phase 1's mock
// payment flow the order is already 'paid' almost immediately, but this
// still exercises the real polling logic needed once Phase 2's gateway
// makes payment genuinely asynchronous.
export function OrderConfirmation({ orderId }: { orderId: string }) {
  const startRef = useRef(Date.now())
  const [timedOut, setTimedOut] = useState(false)

  const { data: order } = useQuery({
    queryKey: ['booking', orderId],
    queryFn: () => getBooking(orderId),
    refetchInterval: (query) => {
      const status = query.state.data?.status
      if (status && status !== 'pending') return false
      if (Date.now() - startRef.current > POLL_TIMEOUT_MS) return false
      return POLL_INTERVAL_MS
    },
  })

  useEffect(() => {
    if (order?.status !== 'pending') return
    const timer = setTimeout(() => setTimedOut(true), POLL_TIMEOUT_MS)
    return () => clearTimeout(timer)
  }, [order])

  return (
    <>
      <Header variant="public" />
      <div className="mx-auto max-w-xl px-4 py-16 text-center">
        {!order ? (
          <p className="text-sm text-gray-500">Đang tải...</p>
        ) : order.status === 'paid' ? (
          <>
            <CircleCheck className="mx-auto h-12 w-12 text-green-600" />
            <h1 className="mt-4 text-2xl font-bold text-gray-900">Đặt vé thành công!</h1>
            <p className="mt-2 text-gray-500">
              Tổng cộng {formatCurrency(order.total_amount)} — {order.tickets.length} vé đã được phát hành.
            </p>
            <Link href={`/me/bookings/${order.id}`} className="mt-6 inline-block">
              <Button>Xem vé của tôi</Button>
            </Link>
          </>
        ) : order.status === 'pending' && !timedOut ? (
          <>
            <Clock className="mx-auto h-12 w-12 animate-pulse text-amber-500" />
            <h1 className="mt-4 text-2xl font-bold text-gray-900">Đang xử lý thanh toán...</h1>
            <p className="mt-2 text-gray-500">Vui lòng đợi trong giây lát.</p>
          </>
        ) : (
          <>
            <CircleX className="mx-auto h-12 w-12 text-red-600" />
            <h1 className="mt-4 text-2xl font-bold text-gray-900">
              {order.status === 'pending' ? 'Quá thời gian xử lý' : 'Đơn hàng không thành công'}
            </h1>
            <p className="mt-2 text-gray-500">Vui lòng kiểm tra lại đơn hàng trong mục Vé của tôi.</p>
            <Link href="/me/bookings" className="mt-6 inline-block">
              <Button variant="secondary">Vé của tôi</Button>
            </Link>
          </>
        )}
      </div>
    </>
  )
}
