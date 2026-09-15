// Hand-written interfaces mirroring api-docs/openapi/{identity,event,booking}-service.yaml
// response schemas (see docs/06-frontend/README.md's Zod-mirrors-OpenAPI rule —
// request bodies are Zod-validated in lib/schemas/, these are for response typing).

export type Role = 'super_admin' | 'organizer' | 'user'
export type UserStatus = 'active' | 'pending' | 'banned'

export interface User {
  id: string
  email: string
  full_name: string | null
  avatar_url: string | null
  role: Role
  status: UserStatus
  created_at: string
}

export type Category = 'concert' | 'workshop' | 'sport'
export type EventStatus = 'draft' | 'published' | 'cancelled'

export interface TicketType {
  id: string
  name: string
  price: number
  currency: string
  quota: number
  sold_count: number
}

export interface EventSummary {
  id: string
  title: string
  slug: string
  category: Category
  city: string | null
  start_time: string
  banner_url: string | null
  min_price: number | null
}

export interface EventDetail extends EventSummary {
  organizer_id: string
  venue_name: string | null
  address: string | null
  end_time: string | null
  description: string | null
  status: EventStatus
  ticket_types: TicketType[]
  attributes: Record<string, unknown>
  tags: string[]
}

export type OrderStatus = 'pending' | 'paid' | 'cancelled' | 'expired'
export type TicketStatus = 'valid' | 'used' | 'cancelled'

export interface OrderItem {
  id: string
  ticket_type_id: string
  quantity: number
  unit_price: number
}

export interface Ticket {
  id: string
  order_item_id: string
  ticket_code: string
  status: TicketStatus
  issued_at: string
}

export interface Order {
  id: string
  user_id: string
  status: OrderStatus
  total_amount: number
  expires_at: string
  created_at: string
  items: OrderItem[]
  tickets: Ticket[]
}

export interface Paginated<T> {
  items: T[]
  total: number
}
