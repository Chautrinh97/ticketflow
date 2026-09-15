// Chrome-less layout (no Header/Footer) for /login, per auth-login.md.
export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">{children}</div>
}
