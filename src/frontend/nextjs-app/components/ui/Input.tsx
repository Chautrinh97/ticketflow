'use client'

import { Eye, EyeOff } from 'lucide-react'
import { forwardRef, useState, type InputHTMLAttributes, type TextareaHTMLAttributes } from 'react'

import { InlineFormError } from '@/components/feedback/InlineFormError'
import { cn } from '@/lib/utils/cn'

interface BaseProps {
  label?: string
  error?: string
  helperText?: string
}

type InputProps = BaseProps & InputHTMLAttributes<HTMLInputElement>

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { label, error, helperText, className, type = 'text', id, ...props },
  ref
) {
  const [showPassword, setShowPassword] = useState(false)
  const isPassword = type === 'password'
  const inputType = isPassword && showPassword ? 'text' : type

  return (
    <div className="w-full">
      {label && (
        <label htmlFor={id} className="mb-1 block text-sm font-medium text-gray-700">
          {label}
        </label>
      )}
      <div className="relative">
        <input
          ref={ref}
          id={id}
          type={inputType}
          className={cn(
            'w-full rounded-md border px-4 py-2 text-sm',
            'focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2',
            error ? 'border-red-500' : 'border-gray-200',
            props.disabled && 'bg-gray-50 text-gray-400',
            isPassword && 'pr-10',
            className
          )}
          {...props}
        />
        {isPassword && (
          <button
            type="button"
            aria-label={showPassword ? 'Ẩn mật khẩu' : 'Hiện mật khẩu'}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400"
            onClick={() => setShowPassword((v) => !v)}
          >
            {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        )}
      </div>
      {error ? (
        <InlineFormError message={error} />
      ) : helperText ? (
        <p className="mt-1 text-xs text-gray-400">{helperText}</p>
      ) : null}
    </div>
  )
})

type TextareaProps = BaseProps & TextareaHTMLAttributes<HTMLTextAreaElement>

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(function Textarea(
  { label, error, helperText, className, id, ...props },
  ref
) {
  return (
    <div className="w-full">
      {label && (
        <label htmlFor={id} className="mb-1 block text-sm font-medium text-gray-700">
          {label}
        </label>
      )}
      <textarea
        ref={ref}
        id={id}
        className={cn(
          'w-full resize-y rounded-md border px-4 py-2 text-sm',
          'focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2',
          error ? 'border-red-500' : 'border-gray-200',
          className
        )}
        {...props}
      />
      {error ? (
        <InlineFormError message={error} />
      ) : helperText ? (
        <p className="mt-1 text-xs text-gray-400">{helperText}</p>
      ) : null}
    </div>
  )
})
