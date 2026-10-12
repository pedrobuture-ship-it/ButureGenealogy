import { useState, type FormEvent } from 'react'
import { Field } from '../../components/Field'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAsync } from '../../hooks/useAsync'
import type { SemenLot } from '../../types/domain'
import { api, errorMessage, fieldErrors } from '../../utils/api'
import { entryReasonLabel } from '../../utils/labels'

interface Props {
  bullId: string
  lot?: SemenLot // editing when present
  onClose: () => void
  onSaved: () => void
}

/** Creates a lot (supplier, batch, type, price, initial doses) or edits its descriptive fields. */
export function LotModal({ bullId, lot, onClose, onSaved }: Props) {
  const toast = useToast()
  const types = useAsync(() => api.listSemenTypes())
  const [f, setF] = useState({
    supplier: lot?.supplier ?? '', batch: lot?.batch ?? '', quantity: '', unitCost: lot?.unitCost != null ? String(lot.unitCost).replace('.', ',') : '',
    semenType: lot?.semenType ?? 'conventional', notes: lot?.notes ?? '', reason: 'purchase',
  })
  const [newType, setNewType] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const set = (k: keyof typeof f) => (v: string) => setF((s) => ({ ...s, [k]: v }))

  async function addType() {
    try {
      const t = await api.createSemenType(newType)
      setNewType('')
      types.reload()
      set('semenType')(t.code)
      toast('success', 'Tipo de sêmen criado.')
    } catch (e) {
      setErrors({ newType: fieldErrors(e).name ?? errorMessage(e) })
    }
  }

  async function submit(e: FormEvent) {
    e.preventDefault()
    const cost = f.unitCost.trim() === '' ? null : Number(f.unitCost.replace(',', '.'))
    if (cost !== null && Number.isNaN(cost)) { setErrors({ unitCost: 'Informe um valor numérico.' }); return }
    const qty = f.quantity.trim() === '' ? 0 : Number(f.quantity)
    if (!lot && (!Number.isInteger(qty) || qty < 0)) { setErrors({ quantity: 'Informe um número inteiro (0 ou mais).' }); return }
    setSaving(true); setErrors({}); setFormError(null)
    try {
      if (lot) await api.updateSemenLot(lot.id, { supplier: f.supplier, batch: f.batch, unitCost: cost, semenType: f.semenType, notes: f.notes })
      else await api.addSemenLot({ bullId, supplier: f.supplier, batch: f.batch, quantity: qty, unitCost: cost, semenType: f.semenType, notes: f.notes, reason: f.reason })
      toast('success', lot ? 'Lote atualizado.' : 'Lote cadastrado.')
      onSaved()
    } catch (err) {
      setErrors(fieldErrors(err)); setFormError(errorMessage(err)); setSaving(false)
    }
  }

  return (
    <Modal title={lot ? 'Editar lote' : 'Novo lote de sêmen'} onClose={onClose} wide
      footer={<>
        <button type="button" className="btn" onClick={onClose} disabled={saving}>Cancelar</button>
        <button type="submit" form="lot-form" className="btn btn-primary" disabled={saving}>{saving ? 'Salvando…' : 'Salvar'}</button>
      </>}>
      <form id="lot-form" onSubmit={submit} noValidate className="form-grid">
        {formError && <p role="alert" className="notice notice-error span-2">{formError}</p>}
        <Field label="Fornecedor" optional>{(id) => <input id={id} autoFocus value={f.supplier} onChange={(e) => set('supplier')(e.target.value)} />}</Field>
        <Field label="Lote" optional>{(id) => <input id={id} value={f.batch} onChange={(e) => set('batch')(e.target.value)} />}</Field>
        <Field label="Tipo de sêmen" error={errors.semenType}>{(id, d) => (
          <select id={id} aria-describedby={d} value={f.semenType} onChange={(e) => set('semenType')(e.target.value)}>
            {(types.data ?? [{ code: 'conventional', name: 'Convencional', isBuiltin: true }]).map((t) => <option key={t.code} value={t.code}>{t.name}</option>)}
          </select>)}
        </Field>
        <Field label="Preço por dose (R$)" optional error={errors.unitCost}>{(id, d) => <input id={id} aria-describedby={d} inputMode="decimal" value={f.unitCost} onChange={(e) => set('unitCost')(e.target.value)} />}</Field>
        {!lot && (
          <>
            <Field label="Doses iniciais" optional error={errors.quantity}>{(id, d) => <input id={id} aria-describedby={d} inputMode="numeric" value={f.quantity} onChange={(e) => set('quantity')(e.target.value)} />}</Field>
            <Field label="Motivo da entrada">{(id) => (
              <select id={id} value={f.reason} onChange={(e) => set('reason')(e.target.value)}>
                {Object.entries(entryReasonLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
              </select>)}
            </Field>
          </>
        )}
        <div className="span-2 row-gap">
          <Field label="Novo tipo de sêmen" optional error={errors.newType}>{(id, d) => <input id={id} aria-describedby={d} value={newType} onChange={(e) => setNewType(e.target.value)} placeholder="ex.: Sexado fêmea" />}</Field>
          <button type="button" className="btn" onClick={addType} disabled={!newType.trim()}>Criar tipo</button>
        </div>
        <div className="span-2"><Field label="Observações" optional>{(id) => <textarea id={id} rows={2} value={f.notes} onChange={(e) => set('notes')(e.target.value)} />}</Field></div>
      </form>
    </Modal>
  )
}
