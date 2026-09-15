import { Header } from '@/components/layout/Header'
import { RequireAuth } from '@/lib/auth/RequireAuth'

// Phase 1 simplification vs. components/layout.md's full dashboard chrome:
// no Sidebar — Phase 1 has exactly one organizer flow (create -> manage
// one event), so a Sidebar with "Tổng quan"/"Sự kiện của tôi" links to
// screens that don't exist yet would be speculative Phase 2 UI. Header
// variant=dashboard is kept for the "Về trang chính"/logout affordance.
export default function OrganizerLayout({ children }: { children: React.ReactNode }) {
  return (
    <RequireAuth role="organizer">
      <Header variant="dashboard" />
      <main className="mx-auto max-w-4xl px-6 py-8">{children}</main>
    </RequireAuth>
  )
}
