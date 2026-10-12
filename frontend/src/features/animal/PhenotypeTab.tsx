import { useMemo, useState, type FormEvent } from 'react'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { EmptyState } from '../../components/EmptyState'
import { Field } from '../../components/Field'
import { useToast } from '../../components/Toast'
import { useAsync } from '../../hooks/useAsync'
import type { Trait, TraitItem } from '../../types/domain'
import { api, errorMessage, fieldErrors } from '../../utils/api'
import { formatDate, formatNumber, localToday } from '../../utils/format'
import { TraitFormModal } from '../traits/TraitFormModal'
import { TraitGroups } from './TraitGroups'


export function PhenotypeTab({ animalId }: { animalId: string }) {
  const toast = useToast()
  const traits = useAsync(() => api.getAnimalTraits(animalId))
  const [adding, setAdding] = useState(false)
  const [toDelete, setToDelete] = useState<TraitItem | null>(null)

  if (traits.loading && !traits.data) return <p role="status" className="muted">Carregando medições…</p>
  if (traits.error) return <div role="alert" className="notice notice-error">{traits.error}</div>
  const groups = traits.data ?? []
  const hasAny = groups.some((g) => g.items.some((i) => i.source === 'measurement'))

  return (
    <>
      <div className="tab-actions">
        <button className="btn btn-primary" onClick={() => setAdding(true)}>Registrar medição</button>
      </div>
      {adding && <MeasurementForm animalId={animalId} onCancel={() => setAdding(false)}
        onSaved={() => { setAdding(false); toast('success', 'Medição registrada.'); traits.reload() }} />}
      {!hasAny ? (
        <EmptyState title="Sem medições fenotípicas">Registre pesos, alturas e outras medições observadas neste animal. Fenótipo é diferente de mérito genético.</EmptyState>
      ) : (
        <TraitGroups groups={groups} source="measurement"
          rowActions={(i) => <button className="btn btn-link" onClick={() => setToDelete(i)}>Excluir</button>} />
      )}
      {toDelete && (
        <ConfirmDialog title="Excluir medição" confirmLabel="Excluir" danger onClose={() => setToDelete(null)}
          onConfirm={async () => { await api.deleteMeasurement(toDelete.id); toast('success', 'Medição excluída.'); traits.reload() }}>
          Excluir a medição de {toDelete.traitName} ({formatNumber(toDelete.value)} {toDelete.unit}) de {formatDate(toDelete.provenance.date)}?
        </ConfirmDialog>
      )}
    </>
  )
}

function MeasurementForm({ animalId, onCancel, onSaved }: { animalId: string; onCancel: () => void; onSaved: () => void }) {
  const catalogue = useAsync<Trait[]>(() => api.listTraits({ kind: 'PHENOTYPE' }))
  const [f, setF] = useState({ traitCode: '', value: '', measuredAt: localToday(), source: '', notes: '' })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [newTrait, setNewTrait] = useState(false)
  const set = (k: keyof typeof f) => (v: string) => setF((s) => ({ ...s, [k]: v }))
  const trait = useMemo(() => catalogue.data?.find((t) => t.code === f.traitCode), [catalogue.data, f.traitCode])

  async function submit(e: FormEvent) {
    e.preventDefault()
    const value = Number(f.value.replace(',', '.'))
    if (f.value.trim() === '' || Number.isNaN(value)) { setErrors({ value: 'Informe um valor numérico.' }); return }
    setSaving(true); setErrors({}); setFormError(null)
    try {
      await api.addMeasurement({ animalId, traitCode: f.traitCode, value, unit: trait?.unit ?? '', measuredAt: f.measuredAt, source: f.source, notes: f.notes })
      onSaved()
    } catch (err) {
      setErrors(fieldErrors(err)); setFormError(errorMessage(err)); setSaving(false)
    }
  }

  return (
    <>
      <form className="card block form-inline" onSubmit={submit} noValidate aria-label="Registrar medição">
        {formError && <p role="alert" className="notice notice-error span-all">{formError}</p>}
        <Field label="Característica" error={errors.traitCode}>
          {(id, d) => (
            <select id={id} aria-describedby={d} value={f.traitCode} onChange={(e) => set('traitCode')(e.target.value)}>
              <option value="">Selecione…</option>
              {(catalogue.data ?? []).map((t) => <option key={t.id} value={t.code}>{t.name}{t.unit ? ` (${t.unit})` : ''}</option>)}
            </select>
          )}
        </Field>
        <Field label="Valor" error={errors.value} hint={trait?.unit ? `Unidade: ${trait.unit}` : undefined}>
          {(id, d) => <input id={id} aria-describedby={d} inputMode="decimal" value={f.value} onChange={(e) => set('value')(e.target.value)} />}
        </Field>
        <Field label="Data" error={errors.measuredAt}>{(id, d) => <input id={id} aria-describedby={d} type="date" value={f.measuredAt} onChange={(e) => set('measuredAt')(e.target.value)} />}</Field>
        <Field label="Origem" optional>{(id) => <input id={id} value={f.source} onChange={(e) => set('source')(e.target.value)} placeholder="balança, técnico…" />}</Field>
        <Field label="Observações" optional>{(id) => <input id={id} value={f.notes} onChange={(e) => set('notes')(e.target.value)} />}</Field>
        <div className="form-actions">
          <button type="button" className="btn btn-link" onClick={() => setNewTrait(true)}>Nova característica…</button>
          <button type="button" className="btn" onClick={onCancel} disabled={saving}>Cancelar</button>
          <button type="submit" className="btn btn-primary" disabled={saving}>{saving ? 'Salvando…' : 'Salvar medição'}</button>
        </div>
      </form>
      {newTrait && <TraitFormModal fixedKind="PHENOTYPE" onClose={() => setNewTrait(false)}
        onSaved={(t) => { setNewTrait(false); catalogue.reload(); set('traitCode')(t.code) }} />}
    </>
  )
}
