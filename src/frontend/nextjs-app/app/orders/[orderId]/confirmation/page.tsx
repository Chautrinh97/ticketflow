'use client'

import { RequireAuth } from '@/lib/auth/RequireAuth'

import { OrderConfirmation } from './OrderConfirmation'

export default function OrderConfirmationPage({ params }: { params: { orderId: string } }) {
  return (
    <RequireAuth>
      <OrderConfirmation orderId={params.orderId} />
    </RequireAuth>
  )
}
