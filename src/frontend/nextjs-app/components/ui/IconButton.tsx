'use client'

import { forwardRef, type ButtonHTMLAttributes } from 'react'

import { cn } from '@/lib/utils/cn'

type Variant = 'ghost' | 'danger-ghost'
type Size = 'sm' | 'md'

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: Size
  'aria-label': string
}

const sizeClasses: Record<Size, string> = { sm: 'h-8 w-8', md: 'h-10 w-10' }
const variantClasses: Record<Variant, string> = {
  ghost: 'text-gray-700 hover:bg-gray-100',
  'danger-ghost': 'text-red-600 hover:bg-red-50',
}

export const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(function IconButton(
  { variant = 'ghost', size = 'md', className, children, ...props },
  ref
) {
  return (
    <button
      ref={ref}
      className={cn(
        'inline-flex items-center justify-center rounded-md transition-colors',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2',
        sizeClasses[size],
        variantClasses[variant],
        className
      )}
      {...props}
    >
      {children}
    </button>
  )
})
