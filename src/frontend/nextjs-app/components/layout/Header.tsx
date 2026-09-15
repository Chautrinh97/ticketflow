'use client'

import { Menu } from 'lucide-react'
import Link from 'next/link'
import { useState } from 'react'

import { Avatar } from '@/components/data-display/Avatar'
import { Button } from '@/components/ui/Button'
import { SearchBar } from '@/components/ui/SearchBar'
import { useAuth } from '@/lib/auth/AuthContext'

// Phase 1 simplification vs. the full components/layout.md spec: no
// NotificationBell (no backing endpoint in scope), no "Trở thành
// Organizer"/"Quản trị hệ thống" menu items (organizer-request and admin
// screens are Phase 2) — see the frontend implementation plan's resolved
// gaps.
export function Header({ variant = 'public' }: { variant?: 'public' | 'dashboard' }) {
  const { user, logout } = useAuth()
  const [menuOpen, setMenuOpen] = useState(false)

  return (
    <header className="sticky top-0 z-40 border-b border-gray-200 bg-white">
      <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3">
        <Link href="/" className="text-lg font-bold text-blue-600">
          TicketFlow
        </Link>

        {variant === 'public' && (
          <>
            <nav className="hidden md:block">
              <Link href="/events" className="text-sm font-medium text-gray-700 hover:text-blue-600">
                Sự kiện
              </Link>
            </nav>
            <div className="flex items-center gap-4">
              <div className="hidden md:block">
                <SearchBar compact />
              </div>
              {user ? (
                <div className="relative">
                  <button onClick={() => setMenuOpen((v) => !v)} aria-label="Menu tài khoản">
                    <Avatar name={user.full_name ?? user.email} avatarUrl={user.avatar_url} size="md" />
                  </button>
                  {menuOpen && (
                    <>
                      <div className="fixed inset-0 z-40" onClick={() => setMenuOpen(false)} />
                      <div className="absolute right-0 z-50 mt-2 w-48 rounded-lg border border-gray-200 bg-white py-1 shadow-xl">
                        <Link
                          href="/me/profile"
                          className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                          onClick={() => setMenuOpen(false)}
                        >
                          Hồ sơ của tôi
                        </Link>
                        <Link
                          href="/me/bookings"
                          className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                          onClick={() => setMenuOpen(false)}
                        >
                          Vé của tôi
                        </Link>
                        {(user.role === 'organizer' || user.role === 'super_admin') && (
                          <Link
                            href="/organizer/events/new"
                            className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                            onClick={() => setMenuOpen(false)}
                          >
                            Tạo sự kiện
                          </Link>
                        )}
                        <button
                          onClick={() => {
                            setMenuOpen(false)
                            logout()
                          }}
                          className="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50"
                        >
                          Đăng xuất
                        </button>
                      </div>
                    </>
                  )}
                </div>
              ) : (
                <Link href="/login">
                  <Button variant="primary" size="sm">
                    Đăng nhập
                  </Button>
                </Link>
              )}
              <button className="md:hidden" aria-label="Menu">
                <Menu className="h-5 w-5" />
              </button>
            </div>
          </>
        )}

        {variant === 'dashboard' && (
          <div className="flex items-center gap-4">
            <Link href="/" className="text-sm text-gray-500 hover:text-gray-700">
              Về trang chính
            </Link>
            {user && <Avatar name={user.full_name ?? user.email} avatarUrl={user.avatar_url} size="md" />}
            <Button variant="ghost" size="sm" onClick={() => logout()}>
              Đăng xuất
            </Button>
          </div>
        )}
      </div>
    </header>
  )
}
