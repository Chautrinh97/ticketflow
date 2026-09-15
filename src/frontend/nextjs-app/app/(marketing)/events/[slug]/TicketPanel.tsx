'use client'

import Link from 'next/link'

import { TicketTypeRow } from '@/components/data-display/TicketTypeRow'
import { Button } from '@/components/ui/Button'
import { useAuth } from '@/lib/auth/AuthContext'
import type { EventDetail } from '@/types/api'

// The "Đặt vé ngay" click reads auth state client-side (per the screen
// spec) and routes to /login (with a redirect back) or straight to
// checkout — RSC output above stays public/cacheable, this is the only
// auth-aware piece on the page.
export function TicketPanel({ event }: { event: EventDetail }) {
  const { user } = useAuth()
  const hasTicketTypes = event.ticket_types.length > 0
  const allSoldOut = hasTicketTypes && event.ticket_types.every((tt) => tt.sold_count >= tt.quota)
  const checkoutPath = `/events/${event.slug}/checkout`
  const href = user ? checkoutPath : `/login?redirect=${encodeURIComponent(checkoutPath)}`

  return (
    <div className="rounded-lg border border-gray-200 p-6">
      <h2 className="text-lg font-semibold text-gray-900">Loại vé</h2>
      {!hasTicketTypes ? (
        <p className="mt-4 text-sm text-gray-500">Sự kiện chưa mở bán vé.</p>
      ) : (
        <div className="mt-4">
          {event.ticket_types.map((tt) => (
            <TicketTypeRow key={tt.id} ticketType={tt} />
          ))}
        </div>
      )}
      <Link href={href} className="mt-6 block">
        <Button className="w-full" disabled={!hasTicketTypes || allSoldOut}>
          {allSoldOut ? 'Đã hết vé' : 'Đặt vé ngay'}
        </Button>
      </Link>
    </div>
  )
}
