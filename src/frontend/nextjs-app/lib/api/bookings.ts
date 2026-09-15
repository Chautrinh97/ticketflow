import { apiFetch } from './client'
import type { Order, Paginated } from '@/types/api'

export interface BookingItemInput {
  ticket_type_id: string
  quantity: number
}

// POST /bookings takes items[] (multi ticket-type per order) — see the
// implementation plan's resolved checkout-cart-scope decision and
// api-docs/openapi/booking-service.yaml.
export function createBooking(items: BookingItemInput[]) {
  return apiFetch<Order>('/bookings', { method: 'POST', body: JSON.stringify({ items }) })
}

export function getBooking(id: string) {
  return apiFetch<Order>(`/bookings/${id}`)
}

export function listMyBookings(params: { status?: string; page?: number; page_size?: number } = {}) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v) qs.set(k, String(v))
  })
  const s = qs.toString()
  return apiFetch<Paginated<Order>>(`/users/me/bookings${s ? `?${s}` : ''}`)
}
