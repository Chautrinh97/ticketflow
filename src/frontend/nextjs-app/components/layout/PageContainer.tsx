import { cn } from '@/lib/utils/cn'

interface PageContainerProps {
  variant?: 'public' | 'dashboard'
  className?: string
  children: React.ReactNode
}

export function PageContainer({ variant = 'public', className, children }: PageContainerProps) {
  return (
    <div className={cn(variant === 'public' ? 'mx-auto max-w-6xl px-4' : 'max-w-full px-6', className)}>{children}</div>
  )
}
