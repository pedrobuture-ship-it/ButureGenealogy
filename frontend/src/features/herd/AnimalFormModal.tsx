import { useEffect, useState, type FormEvent } from 'react'
import { Field } from '../../components/Field'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import type { Animal, AnimalInput, Breed, Farm } from '../../types/domain'
import { api, errorMessage, fieldErrors } from '../../utils/api'
import { categoriesForSex, categoryLabel, statusLabel } from '../../utils/labels'
import { animalTitle } from '../../utils/format'

interface Props {
  farms: Farm[]
  breeds: Breed[]
  animal?: Animal // editing when present
  defaultFarmId?: string
  defaults?: Partial<AnimalInput>
  onClose: () => void
  onSaved: (a: Animal) => void
}

export function AnimalFormModal({ farms, breeds, animal, defaultFarmId, defaults, onClose, onSaved }: Props) {
  const toast = useToast()
  const [form, setForm] = useState<AnimalInput>(() => ({
    farmId: animal?.farmId ?? defaultFarmId ?? farms[0]?.id ?? '',
    registration: animal?.registration ?? '',
    name: animal?.name ?? '',
    sex: animal?.sex ?? 'female',
    breedId: animal?.breedId ?? '',
    birthDate: animal?.birthDate ?? '',
    category: animal?.category ?? 'cow',
    status: animal?.status ?? 'active',
    fatherId: animal?.fatherId ?? '',
    motherId: animal?.motherId ?? '',
    notes: animal?.notes ?? '',
    ...defaults,
  }))
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [sires, setSires] = useState<Animal[]>([])
  const [dams, setDams] = useState<Animal[]>([])

  const set = <K extends keyof AnimalInput>(k: K, v: AnimalInput[K]) => setForm((f) => ({ ...f, [k]: v }))

  // Candidate parents come from the same farm; the backend still enforces the rules.
  useEffect(() => {
    if (!form.farmId) return
    let cancelled = false
    const base = { farmId: form.farmId, includeArchived: true, pageSize: 500, sortBy: 'name' }
    Promise.all([api.listAnimals({ ...base, sex: 'male' }), api.listAnimals({ ...base, sex: 'female' })])
      .then(([m, f]) => {
        if (cancelled) return
        setSires(m.items.filter((a) => a.id !== animal?.id))
        setDams(f.items.filter((a) => a.id !== animal?.id))
      })
      .catch(() => { if (!cancelled) setFormError('Não foi possível carregar a lista de pais e mães.') })
    return () => { cancelled = true }
  }, [form.farmId, animal?.id])

  function changeSex(sex: 'male' | 'female') {
    setForm((f) => ({ ...f, sex, category: categoriesForSex(sex).includes(f.category) ? f.category : sex === 'male' ? 'bull' : 'cow' }))
  }

  async function submit(e: FormEvent) {
    e.preventDefault()
    setSaving(true)
    setErrors({})
    setFormError(null)
    try {
      const saved = animal ? await api.updateAnimal(animal.id, form) : await api.createAnimal(form)
      toast('success', animal ? 'Animal atualizado.' : 'Animal cadastrado.')
      onSaved(saved)
    } catch (err) {
      setErrors(fieldErrors(err))
      setFormError(errorMessage(err))
      setSaving(false)
    }
  }

  return (
    <Modal
      title={animal ? 'Editar animal' : 'Novo animal'}
      onClose={onClose}
      wide
      footer={
        <>
          <button type="button" className="btn" onClick={onClose} disabled={saving}>Cancelar</button>
          <button type="submit" form="animal-form" className="btn btn-primary" disabled={saving}>
            {saving ? 'Salvando…' : 'Salvar'}
          </button>
        </>
      }
    >
      <form id="animal-form" onSubmit={submit} className="form-grid" noValidate>
        {formError && <p role="alert" className="notice notice-error span-2">{formError}</p>}
        <Field label="Propriedade" error={errors.farmId}>
          {(id, d) => (
            <select id={id} aria-describedby={d} value={form.farmId} onChange={(e) => set('farmId', e.target.value)}>
              {farms.map((f) => <option key={f.id} value={f.id}>{f.name}</option>)}
            </select>
          )}
        </Field>
        <Field label="Registro" optional error={errors.registration}>
          {(id, d) => <input id={id} aria-describedby={d} value={form.registration} onChange={(e) => set('registration', e.target.value)} />}
        </Field>
        <Field label="Nome" error={errors.name} hint="Informe o nome ou o registro.">
          {(id, d) => <input id={id} aria-describedby={d} value={form.name} onChange={(e) => set('name', e.target.value)} />}
        </Field>
        <Field label="Sexo" error={errors.sex}>
          {(id, d) => (
            <select id={id} aria-describedby={d} value={form.sex} onChange={(e) => changeSex(e.target.value as 'male' | 'female')}>
              <option value="female">Fêmea</option>
              <option value="male">Macho</option>
            </select>
          )}
        </Field>
        <Field label="Raça" optional error={errors.breedId}>
          {(id, d) => (
            <select id={id} aria-describedby={d} value={form.breedId} onChange={(e) => set('breedId', e.target.value)}>
              <option value="">Não informada</option>
              {breeds.map((b) => <option key={b.id} value={b.id}>{b.name}</option>)}
            </select>
          )}
        </Field>
        <Field label="Categoria" error={errors.category}>
          {(id, d) => (
            <select id={id} aria-describedby={d} value={form.category} onChange={(e) => set('category', e.target.value as AnimalInput['category'])}>
              {categoriesForSex(form.sex).map((c) => <option key={c} value={c}>{categoryLabel[c]}</option>)}
            </select>
          )}
        </Field>
        <Field label="Nascimento" optional error={errors.birthDate}>
          {(id, d) => <input id={id} aria-describedby={d} type="date" value={form.birthDate} onChange={(e) => set('birthDate', e.target.value)} />}
        </Field>
        {animal && (
          <Field label="Status" error={errors.status}>
            {(id, d) => (
              <select id={id} aria-describedby={d} value={form.status} onChange={(e) => set('status', e.target.value as AnimalInput['status'])}>
                {Object.entries(statusLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
              </select>
            )}
          </Field>
        )}
        <Field label="Pai" optional error={errors.fatherId}>
          {(id, d) => (
            <select id={id} aria-describedby={d} value={form.fatherId} onChange={(e) => set('fatherId', e.target.value)}>
              <option value="">Não informado</option>
              {sires.map((a) => <option key={a.id} value={a.id}>{animalTitle(a)}</option>)}
            </select>
          )}
        </Field>
        <Field label="Mãe" optional error={errors.motherId}>
          {(id, d) => (
            <select id={id} aria-describedby={d} value={form.motherId} onChange={(e) => set('motherId', e.target.value)}>
              <option value="">Não informada</option>
              {dams.map((a) => <option key={a.id} value={a.id}>{animalTitle(a)}</option>)}
            </select>
          )}
        </Field>
        <div className="span-2">
          <Field label="Observações" optional>
            {(id) => <textarea id={id} rows={3} value={form.notes} onChange={(e) => set('notes', e.target.value)} />}
          </Field>
        </div>
      </form>
    </Modal>
  )
}
