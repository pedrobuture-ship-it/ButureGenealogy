import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Field } from '../../components/Field'
import { Modal } from '../../components/Modal'
import { useToast } from '../../components/Toast'
import { useAsync } from '../../hooks/useAsync'
import type { Breed, PedigreeView as PedigreeViewDto, TreeNode } from '../../types/domain'
import { api, errorMessage, fieldErrors } from '../../utils/api'
import { formatNumber } from '../../utils/format'
import { PedigreeTree } from './PedigreeTree'

const GENERATION_OPTIONS = [2, 3, 4, 5]

export function PedigreeView({ animalId }: { animalId: string }) {
  const navigate = useNavigate()
  const [generations, setGenerations] = useState(3)
  const [creating, setCreating] = useState<TreeNode | null>(null)

  const [data, setData] = useState<PedigreeViewDto | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [tick, setTick] = useState(0)
  const reload = useCallback(() => setTick((t) => t + 1), [])

  // Keep the previous tree on screen while a new depth loads.
  useEffect(() => {
    let cancelled = false
    setLoading(true)
    api.getPedigree(animalId, generations)
      .then((d) => { if (!cancelled) { setData(d); setError(null) } })
      .catch((e: unknown) => { if (!cancelled) setError(errorMessage(e)) })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [animalId, generations, tick])

  const repeatedIds = useMemo(() => new Set(data?.repeated.map((r) => r.id) ?? []), [data])

  if (error && !data) return <div role="alert" className="notice notice-error">{error}</div>
  if (!data) return <p role="status" className="muted">Carregando genealogia…</p>
  const v = data

  return (
    <div className="ped-layout">
      <div className="ped-toolbar">
        <div role="group" aria-label="Gerações exibidas" className="segmented">
          {GENERATION_OPTIONS.map((n) => (
            <button key={n} className={n === generations ? 'seg active' : 'seg'} aria-pressed={n === generations} onClick={() => setGenerations(n)}>
              {n} gerações
            </button>
          ))}
        </div>
        <span className="muted" aria-live="polite">{loading ? 'Atualizando…' : `Profundidade conhecida: ${v.depth} de ${v.generations}`}</span>
      </div>

      {error && <div role="alert" className="notice notice-error">{error}</div>}

      {v.issues.length > 0 && (
        <div role="alert" className="notice notice-error">
          <strong>Inconsistências encontradas no pedigree</strong>
          <ul className="plain-list">{v.issues.map((i, k) => <li key={k}>{i.message}</li>)}</ul>
        </div>
      )}

      {v.tree && (
        <PedigreeTree root={v.tree} repeatedIds={repeatedIds} onOpen={(id) => navigate(`/rebanho/${id}`)} onCreate={setCreating} />
      )}

      <div className="grid">
        <div className="card">
          <h2>Completude do pedigree</h2>
          <table className="data-table compact">
            <thead><tr><th>Geração</th><th>Conhecidos</th><th>%</th></tr></thead>
            <tbody>
              {v.stats.map((s) => (
                <tr key={s.generation}>
                  <td>{s.generation === 1 ? 'Pais' : s.generation === 2 ? 'Avós' : `${s.generation}ª geração`}</td>
                  <td>{s.known} / {s.expected}</td>
                  <td>{formatNumber(s.percent, 0)}%</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="card">
          <h2>Ancestrais repetidos</h2>
          {v.repeated.length === 0 ? (
            <p className="muted">Nenhum ancestral aparece mais de uma vez nas gerações exibidas.</p>
          ) : (
            <ul className="plain-list">
              {v.repeated.map((r) => (
                <li key={r.id}><strong>{r.name || 'Sem nome'}</strong> <span className="muted">em {r.occurrences.join(', ')}</span></li>
              ))}
            </ul>
          )}
          <h2 className="spaced">Endogamia / parentesco</h2>
          <p><span className="badge badge-muted">{v.inbreeding.message || 'não calculado nesta versão'}</span></p>
          <p className="muted small">Ancestrais repetidos indicam consanguinidade possível, mas o coeficiente formal ainda não é calculado.</p>
        </div>
      </div>

      {creating && <MinimalAncestorModal placeholder={creating} onClose={() => setCreating(null)} onSaved={() => { setCreating(null); reload() }} />}
    </div>
  )
}

function MinimalAncestorModal({ placeholder, onClose, onSaved }: { placeholder: TreeNode; onClose: () => void; onSaved: () => void }) {
  const toast = useToast()
  const breeds = useAsync<Breed[]>(() => api.listBreeds())
  const [f, setF] = useState({ name: '', registration: '', breedId: '' })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const role = placeholder.role === 'dam' ? 'dam' : 'sire'

  async function save() {
    setSaving(true); setErrors({}); setFormError(null)
    try {
      await api.createMinimalAncestor({ childId: placeholder.childId, role, ...f })
      toast('success', role === 'sire' ? 'Pai cadastrado.' : 'Mãe cadastrada.')
      onSaved()
    } catch (e) {
      setErrors(fieldErrors(e)); setFormError(errorMessage(e)); setSaving(false)
    }
  }

  return (
    <Modal title={role === 'sire' ? 'Cadastrar pai' : 'Cadastrar mãe'} onClose={onClose}
      footer={<>
        <button className="btn" onClick={onClose} disabled={saving}>Cancelar</button>
        <button className="btn btn-primary" onClick={save} disabled={saving}>{saving ? 'Salvando…' : 'Salvar'}</button>
      </>}>
      <p className="muted">Registro mínimo: informe só o que você sabe. Você pode completar a ficha depois.</p>
      <form className="form-grid" onSubmit={(e) => { e.preventDefault(); void save() }} noValidate>
        {formError && <p role="alert" className="notice notice-error span-2">{formError}</p>}
        <Field label="Nome" error={errors.name}>{(id, d) => <input id={id} aria-describedby={d} autoFocus value={f.name} onChange={(e) => setF({ ...f, name: e.target.value })} />}</Field>
        <Field label="Registro" optional error={errors.registration}>{(id, d) => <input id={id} aria-describedby={d} value={f.registration} onChange={(e) => setF({ ...f, registration: e.target.value })} />}</Field>
        <div className="span-2">
          <Field label="Raça" optional>{(id) => (
            <select id={id} value={f.breedId} onChange={(e) => setF({ ...f, breedId: e.target.value })}>
              <option value="">Não informada</option>
              {(breeds.data ?? []).map((b) => <option key={b.id} value={b.id}>{b.name}</option>)}
            </select>
          )}</Field>
        </div>
      </form>
    </Modal>
  )
}
