import { createContext, useContext } from 'react'

export interface ToastContextValue {
  success: (mensagem: string) => void
  error: (mensagem: string) => void
  info: (mensagem: string) => void
}

export const ToastContext = createContext<ToastContextValue | null>(null)

export function useToasts(): ToastContextValue {
  const ctx = useContext(ToastContext)
  if (ctx === null) throw new Error('useToasts precisa de <ToastProvider>')
  return ctx
}