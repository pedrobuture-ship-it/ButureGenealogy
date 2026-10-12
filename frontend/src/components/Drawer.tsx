import { useEffect, useRef, type ReactNode } from 'react'

interface Props {
  title: string
  onClose: () => void
  children: ReactNode
}

/** Right-hand panel for long, read-only content such as "Ver cálculo". */
export function Drawer({ title, onClose, children }: Props) {
  const ref = useRef<HTMLDivElement>(null)
  const previous = useRef<Element | null>(null)

  useEffect(() => {
    previous.current = document.activeElement
    ref.current?.focus()
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose() }
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('keydown', onKey)
      if (previous.current instanceof HTMLElement) previous.current.focus()
    }
  }, [onClose])

  return (
    <div className="drawer-backdrop" onMouseDown={(e) => e.target === e.currentTarget && onClose()}>
      <aside className="drawer" role="dialog" aria-modal="true" aria-label={title} tabIndex={-1} ref={ref}>
        <header className="drawer-head">
          <h2>{title}</h2>
          <button className="icon-btn" onClick={onClose} aria-label="Fechar">×</button>
        </header>
        <div className="drawer-body">{children}</div>
      </aside>
    </div>
  )
}
