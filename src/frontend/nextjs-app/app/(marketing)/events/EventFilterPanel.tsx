'use client'

import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useState } from 'react'

import { Button } from '@/components/ui/Button'
import { Select } from '@/components/ui/Select'

const categories = [
  { value: 'concert', label: 'Âm nhạc' },
  { value: 'workshop', label: 'Workshop' },
  { value: 'sport', label: 'Thể thao' },
]

interface EventFilterPanelProps {
  initialFilters: { category?: string; city?: string }
}

export function EventFilterPanel({ initialFilters }: EventFilterPanelProps) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const [category, setCategory] = useState(initialFilters.category ?? '')
  const [city, setCity] = useState(initialFilters.city ?? '')

  function apply(nextCategory: string, nextCity: string) {
    const params = new URLSearchParams(searchParams.toString())
    if (nextCategory) params.set('category', nextCategory)
    else params.delete('category')
    if (nextCity) params.set('city', nextCity)
    else params.delete('city')
    router.push(`${pathname}?${params.toString()}`)
  }

  const hasFilters = Boolean(category || city)

  return (
    <div className="flex flex-wrap items-end gap-4">
      <div className="w-48">
        <Select
          label="Thể loại"
          placeholder="Tất cả"
          options={categories}
          value={category}
          onChange={(e) => {
            setCategory(e.target.value)
            apply(e.target.value, city)
          }}
        />
      </div>
      <div className="w-48">
        <label className="mb-1 block text-sm font-medium text-gray-700">Thành phố</label>
        <input
          value={city}
          onChange={(e) => setCity(e.target.value)}
          onBlur={() => apply(category, city)}
          placeholder="Vd: Hà Nội"
          className="w-full rounded-md border border-gray-200 px-4 py-2 text-sm focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
        />
      </div>
      {hasFilters && (
        <Button
          variant="ghost"
          onClick={() => {
            setCategory('')
            setCity('')
            router.push(pathname)
          }}
        >
          Xoá bộ lọc
        </Button>
      )}
    </div>
  )
}
