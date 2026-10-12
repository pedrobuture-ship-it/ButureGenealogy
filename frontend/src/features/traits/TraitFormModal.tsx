import { useState, type FormEvent } from 'react'
import { Field } from '../../components/Field'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import type { Trait, TraitInput } from '../../types/domain'
import { api, errorMessage, fieldErrors } from '../../utils/api'
import { directionLabel, traitCategoryLabel, valueKindLabel } from '../../utils/labels'

interface Props {
  /** When set, the value kind is fixed (e.g. creating a phenotype from the animal page). */
  fixedKind?: string
  onClose: () => void
  onSaved: (t: Trait) => void
}

export function TraitFormModal({ fixedKind, onClose, onSaved }: Props) {
  const toast = useToast()
  const [f, setF] = useState({
    code: '', name: '', description: '', category: 'other', unit: '', dataType: 'number',
    evaluationType: fixedKind ?? 'PHENOTYPE', optimizationDirection: 'INFORMATION_ONLY', idealMin: '', idealMax: '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const set = (k: keyof typeof f) => (v: string) => setF((s) => ({ ...s, [k]: v }))

  async function submit(e: FormEvent) {
    e.preventDefault()
    const num = (s: string) => (s.trim() === '' ? null : Number(s.replace(',', '.')))
    const input: TraitInput = {
      code: f.code, name: f.name, description: f.description, category: f.category, purpose: 'any', unit: f.unit,
      dataType: f.dataType, evaluationType: f.evaluationType, optimizationDirection: f.optimizationDirection,
      idealMin: num(f.idealMin), idealMax: num(f.idealMax), higherIsBetter: null,
    }
    setSaving(true); setErrors({}); setFormError(null)
    try {
      const t = await api.createTrait(input)
      toast('success', 'Característica criada.')
      onSaved(t)
    } catch (err) {
      setErrors(fieldErrors(err)); setFormError(errorMessage(err)); setSaving(false)
    }
  }

  return (
    <Modal title="Nova característica" onClose={onClose} wide
      footer={<>
        <button type="button" className="btn" onClick={onClose} disabled={saving}>Cancelar</button>
        <button type="submit" form="trait-form" className="btn btn-primary" disabled={saving}>{saving ? 'Salvando…' : 'Salvar'}</button>
      </>}>
      <form id="trait-form" className="form-grid" onSubmit={submit} noValidate>
        {formError && <p role="alert" className="notice notice-error span-2">{formError}</p>}
        <Field label="Código" error={errors.code} hint="Letras, números e _ (ex.: BODY_LENGTH)">{(id, d) => <input id={id} aria-describedby={d} autoFocus value={f.code} onChange={(e) => set('code')(e.target.value)} />}</Field>
        <Field label="Nome" error={errors.name}>{(id, d) => <input id={id} aria-describedby={d} value={f.name} onChange={(e) => set('name')(e.target.value)} />}</Field>
        <Field label="Categoria" error={errors.category}>{(id) => (
          <select id={id} value={f.category} onChange={(e) => set('category')(e.target.value)}>
            {Object.entries(traitCategoryLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
          </select>)}
        </Field>
        <Field label="Tipo de valor" error={errors.evaluationType}>{(id) => (
          <select id={id} value={f.evaluationType} disabled={!!fixedKind} onChange={(e) => set('evaluationType')(e.target.value)}>
            {Object.entries(valueKindLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
          </select>)}
        </Field>
        <Field label="Unidade" optional>{(id) => <input id={id} value={f.unit} onChange={(e) => set('unit')(e.target.value)} />}</Field>
        <Field label="Tipo de dado">{(id) => (
          <select id={id} value={f.dataType} onChange={(e) => set('dataType')(e.target.value)}>
            <option value="number">Número</option><option value="integer">Inteiro</option><option value="score">Escore (inteiro)</option>
          </select>)}
        </Field>
        <Field label="Direção de otimização" hint="Pode ficar como apenas informativa; o objetivo é definido nos perfis.">{(id) => (
          <select id={id} value={f.optimizationDirection} onChange={(e) => set('optimizationDirection')(e.target.value)}>
            {Object.entries(directionLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
          </select>)}
        </Field>
        {f.optimizationDirection === 'TARGET_RANGE' && (
          <>
            <Field label="Mínimo ideal" error={errors.idealMin}>{(id, d) => <input id={id} aria-describedby={d} inputMode="decimal" value={f.idealMin} onChange={(e) => set('idealMin')(e.target.value)} />}</Field>
            <Field label="Máximo ideal" error={errors.idealMax}>{(id, d) => <input id={id} aria-describedby={d} inputMode="decimal" value={f.idealMax} onChange={(e) => set('idealMax')(e.target.value)} />}</Field>
          </>
        )}
        <div className="span-2"><Field label="Descrição" optional>{(id) => <textarea id={id} rows={2} value={f.description} onChange={(e) => set('description')(e.target.value)} />}</Field></div>
      </form>
    </Modal>
  )
}
