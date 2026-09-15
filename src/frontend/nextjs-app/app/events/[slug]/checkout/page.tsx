'use client'

import { RequireAuth } from '@/lib/auth/RequireAuth'

import { CheckoutFlow } from './CheckoutFlow'

export default function CheckoutPage({ params }: { params: { slug: string } }) {
  return (
    <RequireAuth>
      <CheckoutFlow slug={params.slug} />
    </RequireAuth>
  )
}
