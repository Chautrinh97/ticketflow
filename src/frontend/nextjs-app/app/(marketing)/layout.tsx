import { Footer } from '@/components/layout/Footer'
import { Header } from '@/components/layout/Header'

export default function MarketingLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <Header variant="public" />
      <main>{children}</main>
      <Footer />
    </>
  )
}
