import { z } from 'zod'

// Mirrors EventCreateInput in api-docs/openapi/event-service.yaml — banner_url
// omitted here since EventCreateInput itself has none (Phase 1 event-form.md
// ships banner-less, see the implementation plan).
export const eventCreateSchema = z.object({
  title: z.string().min(1, 'Vui lòng nhập tên sự kiện'),
  category: z.enum(['concert', 'workshop', 'sport'], {
    errorMap: () => ({ message: 'Vui lòng chọn loại sự kiện' }),
  }),
  venue_name: z.string().optional(),
  address: z.string().optional(),
  city: z.string().optional(),
  start_time: z.string().min(1, 'Vui lòng chọn thời gian bắt đầu'),
  end_time: z.string().optional(),
  description: z.string().optional(),
})
export type EventCreateFormValues = z.infer<typeof eventCreateSchema>

// Mirrors TicketTypeCreateInput.
export const ticketTypeCreateSchema = z.object({
  name: z.string().min(1, 'Vui lòng nhập tên loại vé'),
  price: z.coerce.number().min(0, 'Giá vé không hợp lệ'),
  quota: z.coerce.number().int().min(1, 'Số lượng vé phải lớn hơn 0'),
})
export type TicketTypeCreateFormValues = z.infer<typeof ticketTypeCreateSchema>
