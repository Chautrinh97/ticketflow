import { CircleAlert } from 'lucide-react'

export function InlineFormError({ message }: { message: string }) {
  return (
    <p className="mt-1 flex items-center gap-1 text-xs text-red-600">
      <CircleAlert className="h-4 w-4 shrink-0" />
      {message}
    </p>
  )
}
