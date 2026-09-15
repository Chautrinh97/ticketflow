'use client'

import { zodResolver } from '@hookform/resolvers/zod'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'

import { useToast } from '@/components/feedback/ToastProvider'
import { Button } from '@/components/ui/Button'
import { Input, Textarea } from '@/components/ui/Input'
import { Select } from '@/components/ui/Select'
import { createEvent } from '@/lib/api/events'
import { eventCreateSchema, type EventCreateFormValues } from '@/lib/schemas/event.schema'

const categories = [
  { value: 'concert', label: 'Âm nhạc' },
  { value: 'workshop', label: 'Workshop' },
  { value: 'sport', label: 'Thể thao' },
]

// Create-only in Phase 1 (no edit form — PUT exists on the API but the
// screen spec's "basic create only" scope doesn't need a separate edit UI
// yet). No banner field: EventCreateInput has none, see event.schema.ts.
export default function NewEventPage() {
  const router = useRouter()
  const { showToast } = useToast()
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<EventCreateFormValues>({ resolver: zodResolver(eventCreateSchema) })

  async function onSubmit(values: EventCreateFormValues) {
    try {
      const event = await createEvent({
        ...values,
        start_time: new Date(values.start_time).toISOString(),
        end_time: values.end_time ? new Date(values.end_time).toISOString() : undefined,
      })
      showToast('success', 'Tạo sự kiện thành công')
      router.push(`/organizer/events/${event.id}`)
    } catch {
      showToast('danger', 'Tạo sự kiện thất bại, vui lòng thử lại')
    }
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">Tạo sự kiện mới</h1>
      <form onSubmit={handleSubmit(onSubmit)} className="mt-6 max-w-xl space-y-4">
        <Input label="Tên sự kiện" error={errors.title?.message} {...register('title')} />
        <div>
          <Select label="Loại sự kiện" placeholder="Chọn loại sự kiện" options={categories} {...register('category')} />
          {errors.category && <p className="mt-1 text-xs text-red-600">{errors.category.message}</p>}
        </div>
        <Input label="Địa điểm" {...register('venue_name')} />
        <Input label="Địa chỉ" {...register('address')} />
        <Input label="Thành phố" {...register('city')} />
        <Input
          label="Thời gian bắt đầu"
          type="datetime-local"
          error={errors.start_time?.message}
          {...register('start_time')}
        />
        <Input label="Thời gian kết thúc" type="datetime-local" {...register('end_time')} />
        <Textarea label="Mô tả" rows={4} {...register('description')} />
        <Button type="submit" loading={isSubmitting} className="w-full">
          Tạo sự kiện
        </Button>
      </form>
    </div>
  )
}
