import type { ReactNode } from 'react'

export function Formula({ children, label }: { children: ReactNode; label?: string }) {
  return <pre className="formula" aria-label={label ?? 'Fórmula'}>{children}</pre>
}

export function Example({ children }: { children: ReactNode }) {
  return (
    <div className="doc-example">
      <p className="doc-example-tag">Exemplo sintético</p>
      {children}
    </div>
  )
}

export type DocStatus = 'implemented' | 'partial' | 'planned' | 'not_implemented'

const statusText: Record<DocStatus, string> = {
  implemented: 'Implementado',
  partial: 'Implementado em parte',
  planned: 'Planejado',
  not_implemented: 'Não implementado',
}
const statusClass: Record<DocStatus, string> = { implemented: 'badge-ok', partial: 'badge-info', planned: 'badge-warn', not_implemented: 'badge-muted' }

export function StatusBadge({ status }: { status: DocStatus }) {
  return <span className={`badge ${statusClass[status]}`}>{statusText[status]}</span>
}
