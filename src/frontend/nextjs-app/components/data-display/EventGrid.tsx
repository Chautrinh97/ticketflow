'use client'

import { CalendarSearch } from 'lucide-react'

import { EventCard } from '@/components/data-display/EventCard'
import { EmptyState } from '@/components/feedback/EmptyState'
import { Skeleton } from '@/components/feedback/Skeleton'
import { Button } from '@/components/ui/Button'
import { useEvents } from '@/lib/hooks/useEvents'
import type { EventSummary, Paginated } from '@/types/api'

interface EventGridProps {
  filters: { category?: string; city?: string; from?: string }
  initialData?: Paginated<EventSummary>
}

// Shared by home.md and event-search.md — both use "Xem thêm" pagination.
export function EventGrid({ filters, initialData }: EventGridProps) {
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading } = useEvents(filters, initialData)

  const items = data?.pages.flatMap((p) => p.items) ?? []

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="aspect-[4/5] w-full" />
        ))}
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <EmptyState
        icon={CalendarSearch}
        title="Chưa có sự kiện nào"
        description="Quay lại sau để xem sự kiện mới, hoặc thử bộ lọc khác."
      />
    )
  }

  return (
    <div>
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {items.map((event) => (
          <EventCard key={event.id} event={event} />
        ))}
      </div>
      {hasNextPage && (
        <div className="mt-8 flex justify-center">
          <Button
            variant="secondary"
            className="w-full max-w-xs"
            loading={isFetchingNextPage}
            onClick={() => fetchNextPage()}
          >
            Xem thêm
          </Button>
        </div>
      )}
    </div>
  )
}
