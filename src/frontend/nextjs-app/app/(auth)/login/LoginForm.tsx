'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { useState, type FormEvent } from 'react'

import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useAuth } from '@/lib/auth/AuthContext'
import { buildMockIdToken, FIREBASE_MOCK_MODE, signInWithEmailPassword, signInWithGoogle } from '@/lib/firebase'

export function LoginForm() {
  const { login } = useAuth()
  const router = useRouter()
  const searchParams = useSearchParams()
  const redirect = searchParams.get('redirect') || '/'

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [fullName, setFullName] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      const idToken = FIREBASE_MOCK_MODE
        ? buildMockIdToken(email, fullName || email)
        : await signInWithEmailPassword(email, password)
      await login(idToken)
      router.replace(redirect)
    } catch {
      setError('Đăng nhập thất bại. Vui lòng kiểm tra lại thông tin.')
    } finally {
      setLoading(false)
    }
  }

  async function handleGoogle() {
    setLoading(true)
    setError(null)
    try {
      const idToken = await signInWithGoogle()
      await login(idToken)
      router.replace(redirect)
    } catch {
      setError('Đăng nhập Google thất bại.')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="w-full max-w-sm rounded-xl border border-gray-200 bg-white p-8 shadow-sm">
      <h1 className="text-xl font-bold text-gray-900">Đăng nhập TicketFlow</h1>
      {FIREBASE_MOCK_MODE && (
        <p className="mt-2 text-xs text-amber-600">
          Chế độ dev (mock) — nhập email/họ tên bất kỳ, không cần Firebase thật.
        </p>
      )}
      <form onSubmit={handleSubmit} className="mt-6 space-y-4">
        <Input label="Email" type="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
        {FIREBASE_MOCK_MODE ? (
          <Input label="Họ tên" required value={fullName} onChange={(e) => setFullName(e.target.value)} />
        ) : (
          <Input
            label="Mật khẩu"
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        )}
        {error && <p className="text-xs text-red-600">{error}</p>}
        <Button type="submit" className="w-full" loading={loading}>
          Đăng nhập
        </Button>
      </form>
      {!FIREBASE_MOCK_MODE && (
        <Button variant="secondary" className="mt-3 w-full" onClick={handleGoogle} loading={loading}>
          Đăng nhập với Google
        </Button>
      )}
    </div>
  )
}
