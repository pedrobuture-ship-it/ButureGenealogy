import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react'

type Kind = 'success' | 'error'
interface ToastItem { id: number; kind: Kind; text: string }

const ToastContext = createContext<(kind: Kind, text: string) => void>(() => {})

export const useToast = () => useContext(ToastContext)

export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([])

  const push = useCallback((kind: Kind, text: string) => {
    const id = Date.now() + Math.random()
    setItems((cur) => [...cur, { id, kind, text }])
    setTimeout(() => setItems((cur) => cur.filter((t) => t.id !== id)), kind === 'error' ? 8000 : 4000)
  }, [])

  // A rejected promise nobody handled must still be visible to the user (and in the console).
  useEffect(() => {
    const onRejection = (e: PromiseRejectionEvent) => {
      console.error('Promessa rejeitada sem tratamento', e.reason)
      const reason: unknown = e.reason
      push('error', reason instanceof Error ? reason.message : 'Ocorreu um erro inesperado.')
    }
    window.addEventListener('unhandledrejection', onRejection)
    return () => window.removeEventListener('unhandledrejection', onRejection)
  }, [push])

  return (
    <ToastContext.Provider value={push}>
      {children}
      <div className="toasts" role="status" aria-live="polite">
        {items.map((t) => (
          <div key={t.id} className={`toast toast-${t.kind}`}>{t.text}</div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}
