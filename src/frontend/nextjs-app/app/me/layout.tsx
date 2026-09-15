import { Footer } from '@/components/layout/Footer'
import { Header } from '@/components/layout/Header'
import { RequireAuth } from '@/lib/auth/RequireAuth'

export default function MeLayout({ children }: { children: React.ReactNode }) {
  return (
    <RequireAuth>
      <Header variant="public" />
      <main>{children}</main>
      <Footer />
    </RequireAuth>
  )
}
