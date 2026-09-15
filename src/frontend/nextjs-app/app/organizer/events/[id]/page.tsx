'use client'

import { zodResolver } from '@hookform/resolvers/zod'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'

import { EventStatusBadge } from '@/components/data-display/Badge'
import { TicketTypeRow } from '@/components/data-display/TicketTypeRow'
import { Skeleton } from '@/components/feedback/Skeleton'
import { useToast } from '@/components/feedback/ToastProvider'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { createTicketType, getOrganizerEvent, publishEvent } from '@/lib/api/events'
import { ticketTypeCreateSchema, type TicketTypeCreateFormValues } from '@/lib/schemas/event.schema'
import { formatDateTime } from '@/lib/utils/formatDate'

// Basic tabs only, per phase-1-mvp.md ("Chưa cần dashboard organizer/admin
// đầy đủ"): view detail + add ticket types + publish. No edit/cancel
// buttons, no Người mua vé/Thống kê tabs (those need endpoints/logic
// explicitly deferred to Phase 2). Backed by the Phase 1 addition
// GET /organizer/events/{id} — see lib/api/events.ts.
export default function EventManagePage({ params }: { params: { id: string } }) {
  const { showToast } = useToast()
  const queryClient = useQueryClient()
  const { data: event, isLoading } = useQuery({
    queryKey: ['organizerEvent', params.id],
    queryFn: () => getOrganizerEvent(params.id),
  })

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<TicketTypeCreateFormValues>({ resolver: zodResolver(ticketTypeCreateSchema) })

  async function onAddTicketType(values: TicketTypeCreateFormValues) {
    try {
      await createTicketType(params.id, values)
      reset()
      showToast('success', 'Thêm loại vé thành công')
      queryClient.invalidateQueries({ queryKey: ['organizerEvent', params.id] })
    } catch {
      showToast('danger', 'Thêm loại vé thất bại')
    }
  }

  async function handlePublish() {
    try {
      await publishEvent(params.id)
      showToast('success', 'Đã xuất bản sự kiện')
      queryClient.invalidateQueries({ queryKey: ['organizerEvent', params.id] })
    } catch {
      showToast('danger', 'Xuất bản thất bại')
    }
  }

  if (isLoading || !event) {
    return <Skeleton className="h-64 w-full" />
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">{event.title}</h1>
        <EventStatusBadge status={event.status} />
      </div>
      <p className="mt-1 text-sm text-gray-500">{formatDateTime(event.start_time)}</p>

      <div className="mt-8">
        <h2 className="text-lg font-semibold text-gray-900">Loại vé</h2>
        <div className="mt-4">
          {event.ticket_types.length === 0 ? (
            <p className="text-sm text-gray-500">Chưa có loại vé nào.</p>
          ) : (
            event.ticket_types.map((tt) => (
              <TicketTypeRow key={tt.id} ticketType={tt} soldOverride={`${tt.sold_count}/${tt.quota} đã bán`} />
            ))
          )}
        </div>
        <form onSubmit={handleSubmit(onAddTicketType)} className="mt-6 flex flex-wrap items-start gap-3">
          <div className="w-40">
            <Input placeholder="Tên loại vé" error={errors.name?.message} {...register('name')} />
          </div>
          <div className="w-32">
            <Input placeholder="Giá vé" type="number" error={errors.price?.message} {...register('price')} />
          </div>
          <div className="w-32">
            <Input placeholder="Số lượng" type="number" error={errors.quota?.message} {...register('quota')} />
          </div>
          <Button type="submit" loading={isSubmitting}>
            Thêm loại vé
          </Button>
        </form>
      </div>

      {event.status === 'draft' && (
        <div className="mt-8 border-t border-gray-200 pt-6">
          <Button onClick={handlePublish} disabled={event.ticket_types.length === 0}>
            Xuất bản sự kiện
          </Button>
          {event.ticket_types.length === 0 && (
            <p className="mt-2 text-xs text-gray-400">Cần thêm ít nhất 1 loại vé trước khi xuất bản.</p>
          )}
        </div>
      )}
    </div>
  )
}
