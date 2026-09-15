import { apiFetch } from './client'
import type { User } from '@/types/api'

export function login(firebaseIdToken: string) {
  return apiFetch<{ access_token: string; expires_in: number; user: User }>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ firebase_id_token: firebaseIdToken }),
    skipAuthRetry: true,
  })
}

export function refresh() {
  return apiFetch<{ access_token: string; expires_in: number }>('/auth/refresh', {
    method: 'POST',
    skipAuthRetry: true,
  })
}

export function logout() {
  return apiFetch<void>('/auth/logout', { method: 'POST' })
}

export function getMe() {
  return apiFetch<User>('/users/me', { skipAuthRetry: true })
}

export function updateMe(input: { full_name?: string; avatar_url?: string }) {
  return apiFetch<User>('/users/me', { method: 'PATCH', body: JSON.stringify(input) })
}
