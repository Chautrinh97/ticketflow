'use client'

import { Button } from '@/components/ui/Button'

// CSR runtime error boundary, per docs/06-frontend/screens/system/error-states.md:
// shown where content would appear, with a retry button — never a toast.
export default function Error({ reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return (
    <div className="mx-auto flex min-h-[60vh] max-w-xl flex-col items-center justify-center gap-4 px-4 text-center">
      <h1 className="text-2xl font-bold text-gray-900">Đã có lỗi xảy ra</h1>
      <p className="text-sm text-gray-500">Vui lòng thử lại. Nếu lỗi vẫn tiếp diễn, hãy quay lại sau.</p>
      <Button onClick={reset}>Thử lại</Button>
    </div>
  )
}
