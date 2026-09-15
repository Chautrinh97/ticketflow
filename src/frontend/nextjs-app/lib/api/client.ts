import { tokenStore } from '@/lib/auth/tokenStore'

// NEXT_PUBLIC_* vars are baked into BOTH the browser and server bundles at
// build time with the same value — but in docker-compose, the browser must
// reach the gateway via the host-published port (http://localhost:8080)
// while server-side code (RSC/SSR fetches, running inside the frontend
// container) must reach it via the compose network's service name
// (http://api-gateway:8080). API_BASE_URL_INTERNAL is a plain (non-public)
// env var read at runtime by the Node server for exactly that case; it
// falls back to the public URL for non-docker local dev (`npm run dev`),
// where both contexts really do share http://localhost:8080.
const PUBLIC_API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1'
const INTERNAL_API_BASE = process.env.API_BASE_URL_INTERNAL || PUBLIC_API_BASE

export const API_BASE = typeof window === 'undefined' ? INTERNAL_API_BASE : PUBLIC_API_BASE

export class ApiError extends Error {
  status: number
  code: string
  constructor(status: number, code: string, message: string) {
    super(message)
    this.status = status
    this.code = code
  }
}

interface RequestOptions extends RequestInit {
  // Set internally when retrying after a refresh, and by calls that must
  // never trigger a refresh themselves (login, refresh itself) — prevents
  // infinite recursion.
  skipAuthRetry?: boolean
}

let refreshPromise: Promise<boolean> | null = null

async function refreshAccessToken(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = fetch(`${API_BASE}/auth/refresh`, { method: 'POST', credentials: 'include' })
      .then(async (res) => {
        if (!res.ok) return false
        const data = await res.json()
        tokenStore.set(data.access_token)
        return true
      })
      .catch(() => false)
      .finally(() => {
        refreshPromise = null
      })
  }
  return refreshPromise
}

// apiFetch is the single fetch wrapper every lib/api/*.ts function goes
// through: attaches the bearer token, and on a 401 awaits one shared
// in-flight refresh (guarding against a burst of concurrent 401s each
// firing their own refresh) before retrying the original request once.
// If refresh also fails, clears the token and redirects to /login — per
// docs/06-frontend/README.md's auth rule.
export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const token = tokenStore.get()
  const headers = new Headers(options.headers)
  if (token) headers.set('Authorization', `Bearer ${token}`)
  if (options.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers, credentials: 'include' })

  if (res.status === 401 && !options.skipAuthRetry) {
    const refreshed = await refreshAccessToken()
    if (refreshed) {
      return apiFetch<T>(path, { ...options, skipAuthRetry: true })
    }
    tokenStore.clear()
    if (typeof window !== 'undefined') {
      const redirect = encodeURIComponent(window.location.pathname)
      window.location.assign(`/login?redirect=${redirect}`)
    }
    throw new ApiError(401, 'unauthorized', 'Phiên đăng nhập đã hết hạn')
  }

  if (res.status === 204) return undefined as T

  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new ApiError(res.status, data.code || 'error', data.message || 'Có lỗi xảy ra, vui lòng thử lại')
  }
  return data as T
}
