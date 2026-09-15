import { apiFetch } from './client'

export function createCheckoutSession(orderId: string, provider?: string) {
  return apiFetch<{ payment_id: string; checkout_url: string }>(`/payments/${orderId}/checkout`, {
    method: 'POST',
    body: JSON.stringify(provider ? { provider } : {}),
  })
}
