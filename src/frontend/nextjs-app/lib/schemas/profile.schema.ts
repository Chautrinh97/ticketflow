import { z } from 'zod'

// Mirrors PATCH /users/me's request body — avatar_url omitted from the form
// since Phase 1 has no /files/presign to produce one (see profile.md gap).
export const updateProfileSchema = z.object({
  full_name: z.string().min(1, 'Vui lòng nhập họ tên'),
})
export type UpdateProfileInput = z.infer<typeof updateProfileSchema>
