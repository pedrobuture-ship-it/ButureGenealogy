import { useState, type FormEvent } from 'react'
import { Field } from '../../components/Field'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import type { SemenLot } from '../../types/domain'
import { api, errorMessage, fieldErrors } from '../../utils/api'
import { entryReasonLabel, withdrawalReasonLabel } from '../../utils/labels'

export type MovementKind = 'entry' | 'withdrawal' | 'adjustment'

const titles: Record<MovementKind, string> = { entry: 'Adicionar doses', withdrawal: 'Retirar doses', adjustment: 'Ajustar contagem' }

interface Props {
  kind: MovementKind
  lots: SemenLot[]
  lot?: SemenLot
  onClose: () => void
  onSaved: () => void
}

export function MovementModal({ kind, lots, lot, onClose, onSaved }: Props) {
  const toast = useToast()
  const [lotId, setLotId] = useState(lot?.id ?? lots[0]?.id ?? '')
  const [qty, setQty] = useState('')
  const [reason, setReason] = useState(kind === 'entry' ? 'purchase' : 'insemination')
  const [note, setNote] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const current = lots.find((l) => l.id === lotId)

  async function submit(e: FormEvent) {
    e.preventDefault()
    const n = Number(qty)
    if (!Number.isInteger(n) || (kind === 'adjustment' ? n === 0 : n <= 0)) {
      setErrors({ quantity: kind === 'adjustment' ? 'Informe um número inteiro diferente de zero.' : 'Informe um número inteiro maior que zero.' })
      return
    }
    setSaving(true); setErrors({}); setFormError(null)
    const input = { inventoryId: lotId, quantity: n, reason, note, matingId: '' }
    try {
      if (kind === 'entry') await api.receiveSemen(input)
      else if (kind === 'withdrawal') await api.withdrawSemen(input)
      else await api.adjustSemen(input)
      toast('success', kind === 'entry' ? 'Doses adicionadas.' : kind === 'withdrawal' ? 'Doses retiradas.' : 'Contagem ajustada.')
      onSaved()
    } catch (err) {
      setErrors(fieldErrors(err)); setFormError(errorMessage(err)); setSaving(false)
    }
  }

  const lotLabel = (l: SemenLot) => `${l.supplier || 'Sem fornecedor'}${l.batch ? ` · lote ${l.batch}` : ''} (${l.quantity} doses)`

  return (
    <Modal title={titles[kind]} onClose={onClose}
      footer={<>
        <button type="button" className="btn" onClick={onClose} disabled={saving}>Cancelar</button>
        <button type="submit" form="movement-form" className="btn btn-primary" disabled={saving || !lotId}>{saving ? 'Salvando…' : 'Confirmar'}</button>
      </>}>
      <form id="movement-form" onSubmit={submit} noValidate className="form-grid">
        {formError && <p role="alert" className="notice notice-error span-2">{formError}</p>}
        <div className="span-2">
          <Field label="Lote" error={errors.inventoryId}>{(id) => (
            <select id={id} value={lotId} onChange={(e) => setLotId(e.target.value)}>
              {lots.map((l) => <option key={l.id} value={l.id}>{lotLabel(l)}</option>)}
            </select>)}
          </Field>
        </div>
        <Field label={kind === 'adjustment' ? 'Diferença (use negativo para reduzir)' : 'Quantidade de doses'} error={errors.quantity}
          hint={kind === 'withdrawal' && current ? `Disponível neste lote: ${current.quantity}` : undefined}>
          {(id, d) => <input id={id} aria-describedby={d} inputMode="numeric" autoFocus value={qty} onChange={(e) => setQty(e.target.value)} />}
        </Field>
        {kind !== 'adjustment' && (
          <Field label="Motivo" error={errors.reason}>{(id, d) => (
            <select id={id} aria-describedby={d} value={reason} onChange={(e) => setReason(e.target.value)}>
              {Object.entries(kind === 'entry' ? entryReasonLabel : withdrawalReasonLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
            </select>)}
          </Field>
        )}
        <div className="span-2">
          <Field label="Observação" optional={kind !== 'adjustment'} error={errors.note}>{(id, d) => <input id={id} aria-describedby={d} value={note} onChange={(e) => setNote(e.target.value)} />}</Field>
        </div>
      </form>
    </Modal>
  )
}
