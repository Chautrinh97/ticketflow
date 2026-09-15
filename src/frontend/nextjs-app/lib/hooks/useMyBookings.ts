'use client'

import { useInfiniteQuery } from '@tanstack/react-query'

import { listMyBookings } from '@/lib/api/bookings'

const PAGE_SIZE = 10

// "Xem thêm" pagination, per docs/06-frontend/interaction-patterns.md.
export function useMyBookings(status?: string) {
  return useInfiniteQuery({
    queryKey: ['myBookings', status],
    queryFn: ({ pageParam }) => listMyBookings({ status, page: pageParam, page_size: PAGE_SIZE }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.reduce((sum, p) => sum + p.items.length, 0)
      return loaded < lastPage.total ? allPages.length + 1 : undefined
    },
  })
}
