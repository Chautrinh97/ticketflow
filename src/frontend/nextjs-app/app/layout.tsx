import type { Metadata } from 'next'

import { Providers } from './providers'
import './globals.css'

export const metadata: Metadata = {
  title: 'TicketFlow',
  description: 'Nền tảng đặt vé sự kiện trực tuyến',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="vi">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
