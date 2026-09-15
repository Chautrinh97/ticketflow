'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useEffect } from 'react'

import { Skeleton } from '@/components/feedback/Skeleton'
import { useAuth } from '@/lib/auth/AuthContext'
import type { Role } from '@/types/api'

interface RequireAuthProps {
  children: React.ReactNode
  role?: Role
}

// Client guard for protected CSR routes: shows a skeleton while the
// silent-refresh bootstrap (AuthProvider) resolves, then redirects to
// /login?redirect=<path> if still unauthenticated — per
// docs/06-frontend/README.md's auth rule.
export function RequireAuth({ children, role }: RequireAuthProps) {
  const { user, isLoading } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    if (!isLoading && !user) {
      router.replace(`/login?redirect=${encodeURIComponent(pathname)}`)
    }
  }, [isLoading, user, pathname, router])

  if (isLoading || !user) {
    return (
      <div className="mx-auto max-w-6xl px-4 py-8">
        <Skeleton className="h-40 w-full" />
      </div>
    )
  }

  if (role && user.role !== role && user.role !== 'super_admin') {
    return (
      <div className="mx-auto max-w-6xl px-4 py-16 text-center text-gray-500">
        Bạn không có quyền truy cập trang này.
      </div>
    )
  }

  return <>{children}</>
}
