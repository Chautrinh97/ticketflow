import { z } from 'zod'

// Mirrors POST /bookings's request body (items[]) — api-docs/openapi/booking-service.yaml.
export const bookingItemSchema = z.object({
  ticket_type_id: z.string().uuid(),
  quantity: z.number().int().min(1),
})

export const createBookingSchema = z.object({
  items: z.array(bookingItemSchema).min(1, 'Vui lòng chọn ít nhất 1 vé'),
})
export type CreateBookingInput = z.infer<typeof createBookingSchema>
