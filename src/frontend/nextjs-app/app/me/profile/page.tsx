'use client'

import { useState } from 'react'

import { Avatar } from '@/components/data-display/Avatar'
import { useToast } from '@/components/feedback/ToastProvider'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { updateMe } from '@/lib/api/auth'
import { useAuth } from '@/lib/auth/AuthContext'

// Name edit only in Phase 1 — no avatar upload (no /files/presign backing
// endpoint yet, see profile.md gap in the implementation plan).
export default function ProfilePage() {
  const { user } = useAuth()
  const { showToast } = useToast()
  const [fullName, setFullName] = useState(user?.full_name ?? '')
  const [saving, setSaving] = useState(false)

  if (!user) return null

  const dirty = fullName.trim() !== (user.full_name ?? '') && fullName.trim().length > 0

  async function handleSave() {
    setSaving(true)
    try {
      await updateMe({ full_name: fullName.trim() })
      showToast('success', 'Cập nhật hồ sơ thành công')
    } catch {
      showToast('danger', 'Cập nhật thất bại, vui lòng thử lại')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="mx-auto max-w-xl px-4 py-8">
      <h1 className="text-2xl font-bold text-gray-900">Hồ sơ của tôi</h1>
      <div className="mt-6 flex items-center gap-4">
        <Avatar name={user.full_name ?? user.email} avatarUrl={user.avatar_url} size="lg" />
      </div>
      <div className="mt-6 max-w-sm space-y-4">
        <Input label="Email" value={user.email} disabled readOnly />
        <Input label="Họ tên" value={fullName} onChange={(e) => setFullName(e.target.value)} />
        <Button disabled={!dirty} loading={saving} onClick={handleSave}>
          Lưu thay đổi
        </Button>
      </div>
    </div>
  )
}
