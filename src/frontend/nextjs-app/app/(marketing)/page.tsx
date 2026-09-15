import { EventGrid } from '@/components/data-display/EventGrid'
import { listEvents } from '@/lib/api/events'
import type { EventSummary, Paginated } from '@/types/api'

// ISR per docs/06-frontend/README.md's routing table.
export const revalidate = 60

const EMPTY_PAGE: Paginated<EventSummary> = { items: [], total: 0 }

export default async function HomePage() {
  // Swallow fetch failures here rather than letting them fail the whole
  // page/build: this runs at build time too (ISR prerender), when the
  // backend may not be reachable yet (e.g. building the frontend Docker
  // image before other containers are up) — EventGrid's client-side
  // TanStack Query hook still fetches normally on hydration either way.
  const initialData = await listEvents({ page: 1, page_size: 12 }).catch(() => EMPTY_PAGE)

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <h1 className="text-2xl font-bold text-gray-900">Sự kiện sắp diễn ra</h1>
      <div className="mt-6">
        <EventGrid filters={{}} initialData={initialData} />
      </div>
    </div>
  )
}
