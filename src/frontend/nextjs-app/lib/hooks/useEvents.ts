'use client'

import { useInfiniteQuery } from '@tanstack/react-query'

import { listEvents, type ListEventsParams } from '@/lib/api/events'
import type { Paginated } from '@/types/api'
import type { EventSummary } from '@/types/api'

const PAGE_SIZE = 12

// "Xem thêm" (load-more) pagination per docs/06-frontend/interaction-patterns.md,
// shared by home.md and event-search.md.
export function useEvents(
  filters: Omit<ListEventsParams, 'page' | 'page_size'>,
  initialData?: Paginated<EventSummary>
) {
  return useInfiniteQuery({
    queryKey: ['events', filters],
    queryFn: ({ pageParam }) => listEvents({ ...filters, page: pageParam, page_size: PAGE_SIZE }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((sum, p) => sum + p.items.length, 0)
      return loaded < lastPage.total ? allPages.length + 1 : undefined
    },
    initialData: initialData ? { pages: [initialData], pageParams: [1] } : undefined,
  })
}
