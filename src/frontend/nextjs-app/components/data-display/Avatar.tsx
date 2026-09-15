import { cn } from '@/lib/utils/cn'

const sizeClasses = { sm: 'h-6 w-6 text-xs', md: 'h-8 w-8 text-sm', lg: 'h-20 w-20 text-2xl' }

interface AvatarProps {
  name: string
  avatarUrl?: string | null
  size?: 'sm' | 'md' | 'lg'
}

export function Avatar({ name, avatarUrl, size = 'md' }: AvatarProps) {
  if (avatarUrl) {
    // eslint-disable-next-line @next/next/no-img-element -- external, unoptimized avatar URL
    return <img src={avatarUrl} alt={name} className={cn('rounded-full object-cover', sizeClasses[size])} />
  }
  const initial = name?.trim()?.[0]?.toUpperCase() ?? '?'
  return (
    <div
      className={cn(
        'flex items-center justify-center rounded-full bg-blue-100 font-semibold text-blue-700',
        sizeClasses[size]
      )}
    >
      {initial}
    </div>
  )
}
