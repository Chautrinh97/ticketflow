import { EventGrid } from '@/components/data-display/EventGrid'
import { listEvents } from '@/lib/api/events'

import { EventFilterPanel } from './EventFilterPanel'

interface EventSearchPageProps {
  searchParams: { category?: string; city?: string; from?: string }
}

// SSR on first load (reads searchParams), then CSR "Xem thêm"/filter changes
// via EventGrid's TanStack Query hook — docs/06-frontend/README.md's routing
// table. Phase 1 = basic filters only (category/city/from via GET /events);
// full-text `q` is Phase 2's GET /events/search.
export default async function EventSearchPage({ searchParams }: EventSearchPageProps) {
  const filters = { category: searchParams.category, city: searchParams.city, from: searchParams.from }
  const initialData = await listEvents({ ...filters, page: 1, page_size: 12 })

  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      <h1 className="text-2xl font-bold text-gray-900">Tìm kiếm sự kiện</h1>
      <div className="mt-4">
        <EventFilterPanel initialFilters={filters} />
      </div>
      <div className="mt-6">
        <EventGrid filters={filters} initialData={initialData} />
      </div>
    </div>
  )
}
