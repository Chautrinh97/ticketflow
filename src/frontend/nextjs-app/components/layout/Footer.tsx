export function Footer() {
  return (
    <footer className="border-t border-gray-200 bg-white">
      <div className="mx-auto grid max-w-6xl gap-8 px-4 py-8 md:grid-cols-3">
        <div>
          <p className="text-lg font-bold text-blue-600">TicketFlow</p>
          <p className="mt-2 text-sm text-gray-500">Nền tảng đặt vé sự kiện trực tuyến.</p>
        </div>
        <div>
          <p className="text-sm font-semibold text-gray-900">Liên kết nhanh</p>
          <ul className="mt-2 space-y-1 text-sm text-gray-500">
            <li>Về chúng tôi</li>
            <li>Điều khoản</li>
            <li>Liên hệ</li>
          </ul>
        </div>
        <div>
          <p className="text-sm font-semibold text-gray-900">Mạng xã hội</p>
        </div>
      </div>
      <p className="pb-6 text-center text-xs text-gray-400">© {new Date().getFullYear()} TicketFlow</p>
    </footer>
  )
}
