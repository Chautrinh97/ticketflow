'use client'

import { createContext, useCallback, useContext, useEffect, useState } from 'react'

import * as authApi from '@/lib/api/auth'
import { tokenStore } from '@/lib/auth/tokenStore'
import type { User } from '@/types/api'

interface AuthState {
  user: User | null
  isLoading: boolean
  login: (firebaseIdToken: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthState | null>(null)

// AuthProvider bootstraps a session on mount by calling POST /auth/refresh
// (the browser auto-attaches the httpOnly cookie) — necessary because the
// access token is memory-only and lost on every hard reload/new tab. See
// docs/06-frontend/README.md's auth-token-handling rule.
export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    async function bootstrap() {
      try {
        const { access_token } = await authApi.refresh()
        tokenStore.set(access_token)
        const me = await authApi.getMe()
        if (!cancelled) setUser(me)
      } catch {
        // No valid session cookie — stay logged out, this is the normal
        // state for a first-time/anonymous visitor.
      } finally {
        if (!cancelled) setIsLoading(false)
      }
    }
    bootstrap()

    const unsubscribe = tokenStore.subscribe(() => {
      if (!tokenStore.get()) setUser(null)
    })
    return () => {
      cancelled = true
      unsubscribe()
    }
  }, [])

  const login = useCallback(async (firebaseIdToken: string) => {
    const data = await authApi.login(firebaseIdToken)
    tokenStore.set(data.access_token)
    setUser(data.user)
  }, [])

  const logout = useCallback(async () => {
    try {
      await authApi.logout()
    } finally {
      tokenStore.clear()
      setUser(null)
    }
  }, [])

  return <AuthContext.Provider value={{ user, isLoading, login, logout }}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
