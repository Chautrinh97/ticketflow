import Link from 'next/link'

import { Button } from '@/components/ui/Button'

// SSR 404 — real HTTP status for crawlers, per docs/06-frontend/screens/system/error-states.md.
export default function NotFound() {
  return (
    <div className="mx-auto flex min-h-[60vh] max-w-xl flex-col items-center justify-center gap-4 px-4 text-center">
      <h1 className="text-2xl font-bold text-gray-900">Không tìm thấy trang</h1>
      <p className="text-sm text-gray-500">Trang bạn tìm không tồn tại hoặc đã bị gỡ bỏ.</p>
      <Link href="/">
        <Button>Về trang chủ</Button>
      </Link>
    </div>
  )
}
