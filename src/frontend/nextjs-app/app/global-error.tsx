'use client'

// Root-layout-level crash fallback — must render its own <html>/<body>
// since it replaces RootLayout entirely when the crash happens above it.
export default function GlobalError({ reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return (
    <html lang="vi">
      <body>
        <div className="mx-auto flex min-h-screen max-w-xl flex-col items-center justify-center gap-4 px-4 text-center">
          <h1 className="text-2xl font-bold text-gray-900">Đã có lỗi xảy ra</h1>
          <p className="text-sm text-gray-500">Vui lòng tải lại trang.</p>
          <button
            onClick={reset}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
          >
            Thử lại
          </button>
        </div>
      </body>
    </html>
  )
}
