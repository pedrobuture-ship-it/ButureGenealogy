import { useId, type ReactNode } from 'react'

interface Props {
  label: string
  error?: string
  hint?: string
  optional?: boolean
  children: (id: string, describedBy: string | undefined) => ReactNode
}

/** Labelled form control with inline error and optional hint. */
export function Field({ label, error, hint, optional, children }: Props) {
  const id = useId()
  const descId = error || hint ? `${id}-desc` : undefined
  return (
    <div className={error ? 'field has-error' : 'field'}>
      <label htmlFor={id}>
        {label}
        {optional && <span className="optional"> (opcional)</span>}
      </label>
      {children(id, descId)}
      {(error || hint) && (
        <p id={descId} className={error ? 'field-error' : 'field-hint'} role={error ? 'alert' : undefined}>
          {error || hint}
        </p>
      )}
    </div>
  )
}
