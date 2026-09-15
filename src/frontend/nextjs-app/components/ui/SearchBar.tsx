'use client'

import { Search } from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useState, type FormEvent } from 'react'

interface SearchBarProps {
  compact?: boolean
  initialQuery?: string
}

// Navigates to /events?q=... — Phase 1's event-search.md only wires up
// basic filters (category/city/date) to GET /events; full-text search via
// `q` is Phase 2's GET /events/search (docs/02-domains/event-catalog/spec.md).
// The query box still exists per the Header spec, it just isn't sent to the
// API yet.
export function SearchBar({ compact, initialQuery = '' }: SearchBarProps) {
  const [query, setQuery] = useState(initialQuery)
  const router = useRouter()

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    router.push(query ? `/events?q=${encodeURIComponent(query)}` : '/events')
  }

  return (
    <form onSubmit={handleSubmit} className={compact ? 'w-48 md:w-64' : 'w-full max-w-xl'}>
      <div className="relative">
        <Search className="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Tìm sự kiện, nghệ sĩ, địa điểm..."
          className="w-full rounded-full border border-gray-200 py-2 pl-10 pr-4 text-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
        />
      </div>
    </form>
  )
}
