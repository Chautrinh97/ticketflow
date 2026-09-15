'use client'

import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import { useMemo, useState } from 'react'

import { TicketTypeRow } from '@/components/data-display/TicketTypeRow'
import { useToast } from '@/components/feedback/ToastProvider'
import { Header } from '@/components/layout/Header'
import { Button } from '@/components/ui/Button'
import { createBooking } from '@/lib/api/bookings'
import { ApiError } from '@/lib/api/client'
import { getEventBySlug } from '@/lib/api/events'
import { createCheckoutSession } from '@/lib/api/payments'
import { formatCurrency } from '@/lib/utils/formatCurrency'

type Step = 1 | 2

// checkout.md: 1 page, 2 sequential steps (not 2 routes). Step 1 lets every
// ticket type's QuantityStepper be set independently — one order can mix
// multiple ticket types, per the implementation plan's resolved
// checkout-cart-scope decision (POST /bookings takes items[]).
export function CheckoutFlow({ slug }: { slug: string }) {
  const [step, setStep] = useState<Step>(1)
  const [quantities, setQuantities] = useState<Record<string, number>>({})
  const [submitting, setSubmitting] = useState(false)
  const { showToast } = useToast()
  const router = useRouter()
  const queryClient = useQueryClient()

  // Critical path: always refetch fresh stock right before showing
  // quantities, per docs/06-frontend/README.md's "ticket stock can change
  // while user is on checkout page" concern.
  const { data: event, isLoading } = useQuery({
    queryKey: ['event', slug],
    queryFn: () => getEventBySlug(slug),
    refetchOnMount: 'always',
    staleTime: 0,
  })

  const selectedItems = useMemo(
    () =>
      Object.entries(quantities)
        .filter(([, qty]) => qty > 0)
        .map(([ticket_type_id, quantity]) => ({ ticket_type_id, quantity })),
    [quantities]
  )
  const totalQuantity = selectedItems.reduce((sum, it) => sum + it.quantity, 0)
  const totalAmount = useMemo(() => {
    if (!event) return 0
    return selectedItems.reduce((sum, it) => {
      const tt = event.ticket_types.find((t) => t.id === it.ticket_type_id)
      return sum + (tt ? tt.price * it.quantity : 0)
    }, 0)
  }, [event, selectedItems])

  if (isLoading || !event) {
    return (
      <>
        <Header variant="public" />
        <div className="mx-auto max-w-3xl px-4 py-8 text-sm text-gray-500">Đang tải...</div>
      </>
    )
  }

  // No optimistic update here — must wait for the real POST /bookings
  // response, per docs/06-frontend/interaction-patterns.md#optimistic-update
  // (stock can genuinely run out between selection and submit).
  async function handleSubmit() {
    setSubmitting(true)
    try {
      const order = await createBooking(selectedItems)
      const { checkout_url } = await createCheckoutSession(order.id)
      router.push(checkout_url)
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        showToast('danger', 'Rất tiếc, số lượng vé bạn chọn không còn đủ. Vui lòng chọn lại.')
        setStep(1)
        queryClient.invalidateQueries({ queryKey: ['event', slug] })
      } else if (err instanceof ApiError && err.status === 429) {
        showToast('danger', 'Bạn thao tác quá nhanh, vui lòng thử lại sau vài giây')
      } else {
        showToast('danger', 'Đặt vé thất bại, vui lòng thử lại.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <>
      <Header variant="public" />
      <div className="mx-auto max-w-3xl px-4 py-8 pb-24 md:pb-8">
        <div className="mb-6 flex gap-6 text-sm font-medium">
          <span className={step >= 1 ? 'text-blue-600' : 'text-gray-400'}>1. Chọn vé</span>
          <span className={step >= 2 ? 'text-blue-600' : 'text-gray-400'}>2. Xác nhận</span>
        </div>

        <h1 className="text-xl font-bold text-gray-900">{event.title}</h1>

        {step === 1 && (
          <div className="mt-6">
            {event.ticket_types.map((tt) => (
              <TicketTypeRow
                key={tt.id}
                ticketType={tt}
                quantity={quantities[tt.id] ?? 0}
                onQuantityChange={(value) => setQuantities((prev) => ({ ...prev, [tt.id]: value }))}
              />
            ))}
            <div className="mt-4 flex items-center justify-between border-t border-gray-200 pt-4">
              <span className="text-sm text-gray-500">Tạm tính</span>
              <span className="font-semibold text-gray-900">{formatCurrency(totalAmount)}</span>
            </div>
            <Button className="mt-6 hidden w-full md:flex" disabled={totalQuantity === 0} onClick={() => setStep(2)}>
              Tiếp tục
            </Button>
          </div>
        )}

        {step === 2 && (
          <div className="mt-6">
            <table className="w-full text-sm">
              <tbody>
                {selectedItems.map((item) => {
                  const tt = event.ticket_types.find((t) => t.id === item.ticket_type_id)
                  if (!tt) return null
                  return (
                    <tr key={item.ticket_type_id} className="border-b border-gray-100">
                      <td className="py-2 text-gray-700">{tt.name}</td>
                      <td className="py-2 text-gray-500">x{item.quantity}</td>
                      <td className="py-2 text-right text-gray-700">{formatCurrency(tt.price * item.quantity)}</td>
                    </tr>
                  )
                })}
                <tr>
                  <td className="pt-4 font-semibold text-gray-900" colSpan={2}>
                    Tổng cộng
                  </td>
                  <td className="pt-4 text-right font-semibold text-gray-900">{formatCurrency(totalAmount)}</td>
                </tr>
              </tbody>
            </table>
            <div className="mt-6 flex gap-3">
              <Button variant="secondary" className="flex-1" onClick={() => setStep(1)} disabled={submitting}>
                Đổi lại
              </Button>
              <Button className="flex-1" loading={submitting} onClick={handleSubmit}>
                Tiến hành thanh toán
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* Mobile: fixed total bar instead of scrolling to the bottom, per
          checkout.md's responsive rule. */}
      {step === 1 && (
        <div className="fixed inset-x-0 bottom-0 z-30 border-t border-gray-200 bg-white p-4 md:hidden">
          <div className="mx-auto flex max-w-3xl items-center justify-between gap-4">
            <span className="text-sm text-gray-500">Tạm tính: {formatCurrency(totalAmount)}</span>
            <Button size="sm" disabled={totalQuantity === 0} onClick={() => setStep(2)}>
              Tiếp tục
            </Button>
          </div>
        </div>
      )}
    </>
  )
}
