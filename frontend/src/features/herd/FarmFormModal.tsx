import { useState, type FormEvent } from 'react'
import { Field } from '../../components/Field'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import type { Farm } from '../../types/domain'
import { api, errorMessage, fieldErrors } from '../../utils/api'

export function FarmFormModal({ onClose, onSaved }: { onClose: () => void; onSaved: (f: Farm) => void }) {
  const toast = useToast()
  const [form, setForm] = useState({ name: '', ownerName: '', city: '', state: '', notes: '' })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const set = (k: keyof typeof form) => (v: string) => setForm((f) => ({ ...f, [k]: v }))

  async function submit(e: FormEvent) {
    e.preventDefault()
    setSaving(true)
    setErrors({})
    setFormError(null)
    try {
      const farm = await api.createFarm(form)
      toast('success', 'Propriedade cadastrada.')
      onSaved(farm)
    } catch (err) {
      setErrors(fieldErrors(err))
      setFormError(errorMessage(err))
      setSaving(false)
    }
  }

  return (
    <Modal
      title="Nova propriedade"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn" onClick={onClose} disabled={saving}>Cancelar</button>
          <button type="submit" form="farm-form" className="btn btn-primary" disabled={saving}>
            {saving ? 'Salvando…' : 'Salvar'}
          </button>
        </>
      }
    >
      <form id="farm-form" onSubmit={submit} className="form-grid" noValidate>
        {formError && <p role="alert" className="notice notice-error span-2">{formError}</p>}
        <div className="span-2">
          <Field label="Nome" error={errors.name}>
            {(id, d) => <input id={id} aria-describedby={d} value={form.name} onChange={(e) => set('name')(e.target.value)} autoFocus />}
          </Field>
        </div>
        <Field label="Responsável" optional>
          {(id) => <input id={id} value={form.ownerName} onChange={(e) => set('ownerName')(e.target.value)} />}
        </Field>
        <Field label="Cidade" optional>
          {(id) => <input id={id} value={form.city} onChange={(e) => set('city')(e.target.value)} />}
        </Field>
        <Field label="UF" optional error={errors.state}>
          {(id, d) => <input id={id} aria-describedby={d} maxLength={2} value={form.state} onChange={(e) => set('state')(e.target.value)} />}
        </Field>
        <div className="span-2">
          <Field label="Observações" optional>
            {(id) => <textarea id={id} rows={3} value={form.notes} onChange={(e) => set('notes')(e.target.value)} />}
          </Field>
        </div>
      </form>
    </Modal>
  )
}
