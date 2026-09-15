import { MapPin, Calendar } from 'lucide-react'
import { notFound } from 'next/navigation'

import { ApiError } from '@/lib/api/client'
import { getEventBySlug } from '@/lib/api/events'
import { formatDateTime } from '@/lib/utils/formatDate'

import { TicketPanel } from './TicketPanel'

// SSR, cache:'no-store' — needs freshest ticket stock at request time
// (docs/06-frontend/README.md's routing table), not ISR like home.md.
export default async function EventDetailPage({ params }: { params: { slug: string } }) {
  let event
  try {
    event = await getEventBySlug(params.slug, { cache: 'no-store' })
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound()
    throw err
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <div className="relative aspect-video w-full overflow-hidden rounded-lg bg-gray-100">
        {event.banner_url ? (
          // eslint-disable-next-line @next/next/no-img-element -- manual banner URL, no File Service in Phase 1
          <img src={event.banner_url} alt={event.title} className="h-full w-full object-cover" />
        ) : (
          <div className="flex h-full items-center justify-center text-gray-300">TicketFlow</div>
        )}
      </div>
      <div className="mt-6 grid gap-8 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <h1 className="text-2xl font-bold text-gray-900">{event.title}</h1>
          <div className="mt-2 flex flex-wrap gap-4 text-sm text-gray-500">
            {event.venue_name && (
              <span className="flex items-center gap-1">
                <MapPin className="h-4 w-4" /> {event.venue_name}
                {event.city ? `, ${event.city}` : ''}
              </span>
            )}
            <span className="flex items-center gap-1">
              <Calendar className="h-4 w-4" /> {formatDateTime(event.start_time)}
            </span>
          </div>
          {event.description && <p className="mt-6 whitespace-pre-line text-gray-700">{event.description}</p>}
        </div>
        <div>
          <TicketPanel event={event} />
        </div>
      </div>
    </div>
  )
}
