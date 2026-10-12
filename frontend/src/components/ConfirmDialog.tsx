import { useState, type ReactNode } from 'react'
import { errorMessage } from '../utils/api'
import { Modal } from './Modal'

interface Props {
  title: string
  confirmLabel: string
  danger?: boolean
  onConfirm: () => Promise<void>
  onClose: () => void
  children: ReactNode
}

export function ConfirmDialog({ title, confirmLabel, danger, onConfirm, onClose, children }: Props) {
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function run() {
    setBusy(true)
    setError(null)
    try {
      await onConfirm()
      onClose()
    } catch (e) {
      setError(errorMessage(e))
      setBusy(false)
    }
  }

  return (
    <Modal
      title={title}
      onClose={onClose}
      footer={
        <>
          <button className="btn" onClick={onClose} disabled={busy}>Cancelar</button>
          <button className={danger ? 'btn btn-danger' : 'btn btn-primary'} onClick={run} disabled={busy}>
            {busy ? 'Aguarde…' : confirmLabel}
          </button>
        </>
      }
    >
      <div>{children}</div>
      {error && <p role="alert" className="notice notice-error">{error}</p>}
    </Modal>
  )
}
