// Plain module-level singleton — NOT React state — because lib/api/client.ts's
// fetch wrapper and TanStack Query queryFns need synchronous read/write
// access outside of any component. AuthContext subscribes to it purely to
// trigger re-renders. Access token lives in memory only, per
// docs/06-frontend/README.md: never localStorage/sessionStorage (XSS risk).
type Listener = () => void

let accessToken: string | null = null
const listeners = new Set<Listener>()

function notify() {
  listeners.forEach((l) => l())
}

export const tokenStore = {
  get: () => accessToken,
  set: (token: string) => {
    accessToken = token
    notify()
  },
  clear: () => {
    accessToken = null
    notify()
  },
  subscribe: (listener: Listener) => {
    listeners.add(listener)
    return () => listeners.delete(listener)
  },
}
