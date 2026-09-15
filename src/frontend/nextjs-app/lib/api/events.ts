import { apiFetch } from './client'
import type { EventDetail, EventSummary, Paginated, TicketType } from '@/types/api'

function toQuery<T extends object>(params: T) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== '') qs.set(k, String(v))
  })
  const s = qs.toString()
  return s ? `?${s}` : ''
}

export interface ListEventsParams {
  category?: string
  city?: string
  from?: string
  page?: number
  page_size?: number
}

// GET /events — basic filters only in Phase 1 (category/city/from). The
// full-text/fuzzy `GET /events/search` is Phase 2 (docs/02-domains/event-catalog/spec.md);
// a `q` typed into SearchBar is not sent here yet, see components/ui/SearchBar.tsx.
export function listEvents(params: ListEventsParams = {}) {
  return apiFetch<Paginated<EventSummary>>(`/events${toQuery(params)}`, { skipAuthRetry: true })
}

export function getEventBySlug(slug: string, init: RequestInit = {}) {
  return apiFetch<EventDetail>(`/events/${slug}`, { ...init, skipAuthRetry: true })
}

export function listOrganizerEvents(params: { status?: string; page?: number; page_size?: number } = {}) {
  return apiFetch<Paginated<EventSummary>>(`/organizer/events${toQuery(params)}`)
}

// Phase 1 addition, not in the original OpenAPI file — see
// api-docs/openapi/event-service.yaml and the implementation plan.
export function getOrganizerEvent(id: string) {
  return apiFetch<EventDetail>(`/organizer/events/${id}`)
}

export interface CreateEventInput {
  title: string
  category: string
  venue_name?: string
  address?: string
  city?: string
  start_time: string
  end_time?: string
  description?: string
}

export function createEvent(input: CreateEventInput) {
  return apiFetch<EventDetail>('/organizer/events', { method: 'POST', body: JSON.stringify(input) })
}

export function publishEvent(id: string) {
  return apiFetch<EventDetail>(`/organizer/events/${id}/publish`, { method: 'POST' })
}

export interface CreateTicketTypeInput {
  name: string
  price: number
  currency?: string
  quota: number
}

export function createTicketType(eventId: string, input: CreateTicketTypeInput) {
  return apiFetch<TicketType>(`/organizer/events/${eventId}/ticket-types`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}
