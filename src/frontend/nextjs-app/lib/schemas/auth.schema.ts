import { z } from 'zod'

// Mirrors POST /auth/login's request body ({firebase_id_token}) — the mock
// login form collects email+full_name locally to build that token
// client-side (see lib/firebase.ts), validated here before submit.
export const mockLoginSchema = z.object({
  email: z.string().email('Email không hợp lệ'),
  full_name: z.string().min(1, 'Vui lòng nhập họ tên'),
})
export type MockLoginInput = z.infer<typeof mockLoginSchema>
