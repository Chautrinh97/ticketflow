import { Calendar, MapPin } from 'lucide-react'
import Image from 'next/image'
import Link from 'next/link'

import { formatCurrency } from '@/lib/utils/formatCurrency'
import { formatDate } from '@/lib/utils/formatDate'
import type { EventSummary } from '@/types/api'

const categoryLabel: Record<string, string> = { concert: 'Âm nhạc', workshop: 'Workshop', sport: 'Thể thao' }

export function EventCard({ event }: { event: EventSummary }) {
  return (
    <Link
      href={`/events/${event.slug}`}
      className="group block overflow-hidden rounded-lg border border-gray-200 shadow-sm transition-shadow hover:shadow-md"
    >
      <div className="relative aspect-video bg-gray-100">
        {event.banner_url ? (
          <Image src={event.banner_url} alt={event.title} fill className="object-cover" unoptimized />
        ) : (
          <div className="flex h-full items-center justify-center text-gray-300">TicketFlow</div>
        )}
        <span className="absolute left-2 top-2 inline-flex items-center rounded-md bg-white/90 px-2 py-0.5 text-xs font-medium text-gray-700">
          {categoryLabel[event.category] ?? event.category}
        </span>
      </div>
      <div className="p-4">
        <h3 className="line-clamp-2 text-lg font-semibold text-gray-900">{event.title}</h3>
        <div className="mt-2 flex flex-col gap-1 text-sm text-gray-500">
          {event.city && (
            <span className="flex items-center gap-1">
              <MapPin className="h-4 w-4" /> {event.city}
            </span>
          )}
          <span className="flex items-center gap-1">
            <Calendar className="h-4 w-4" /> {formatDate(event.start_time)}
          </span>
        </div>
        <p className="mt-2 font-semibold text-blue-600">
          {event.min_price != null ? `Từ ${formatCurrency(event.min_price)}` : 'Chưa mở bán vé'}
        </p>
      </div>
    </Link>
  )
}
