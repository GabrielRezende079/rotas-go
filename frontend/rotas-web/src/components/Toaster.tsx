import { useCallback, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { ToastContext } from '../toast'
import type { ToastContextValue } from '../toast'
import { Icon } from './Icon'
import type { IconName } from './Icon'

type TipoToast = 'success' | 'error' | 'info'

interface Toast {
  id: number
  tipo: TipoToast
  mensagem: string
}

const ICONES: Record<TipoToast, IconName> = {
  success: 'check',
  error: 'clear',
  info: 'zap',
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])
  const nextId = useRef(1)

  const dismiss = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const push = useCallback(
    (tipo: TipoToast, mensagem: string, duracao: number) => {
      const id = nextId.current++
      setToasts((prev) => [...prev, { id, tipo, mensagem }])
      window.setTimeout(() => dismiss(id), duracao)
    },
    [dismiss],
  )

  const value: ToastContextValue = {
    success: (m) => push('success', m, 4200),
    error: (m) => push('error', m, 6500),
    info: (m) => push('info', m, 4200),
  }

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="toast-stack" aria-live="polite">
        {toasts.map((toast) => (
          <div key={toast.id} className={`toast toast-${toast.tipo}`}>
            <span className="toast-icon">
              <Icon name={ICONES[toast.tipo]} size={15} />
            </span>
            <span className="toast-mensagem">{toast.mensagem}</span>
            <button
              type="button"
              className="toast-dismiss"
              onClick={() => dismiss(toast.id)}
              aria-label="Fechar aviso"
            >
              <Icon name="close" size={14} />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}