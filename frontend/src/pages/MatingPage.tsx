import { useEffect, useState } from 'react'
import { EmptyState } from '../components/EmptyState'
import { PageHeader } from '../components/PageHeader'
import { useToast } from '../components/Toast'
import { CalculationDrawer } from '../features/mating/CalculationDrawer'
import { CandidatePicker } from '../features/mating/CandidatePicker'
import { FemalePicker } from '../features/mating/FemalePicker'
import { ProfileEditor } from '../features/mating/ProfileEditor'
import { cloneProfile } from '../features/mating/profileUtils'
import { ResultsView } from '../features/mating/ResultsView'
import { useAsync } from '../hooks/useAsync'
import type { Animal } from '../types/domain'
import type { CalculationSummary, Profile, RunResponse } from '../types/engine'
import { api, errorMessage } from '../utils/api'
import { animalTitle, fmt2, formatDateTime } from '../utils/format'

export function MatingPage() {
  const toast = useToast()
  const profiles = useAsync(() => api.listMatingProfiles())
  const [female, setFemale] = useState<Animal | null>(null)
  const [bulls, setBulls] = useState<Set<string>>(new Set())
  const [profile, setProfile] = useState<Profile | null>(null)
  const [running, setRunning] = useState(false)
  const [response, setResponse] = useState<RunResponse | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [view, setView] = useState<null | { matingId?: string; bullId?: string }>(null)

  useEffect(() => {
    if (!profile && profiles.data && profiles.data.length > 0) setProfile(cloneProfile(profiles.data[0]))
  }, [profile, profiles.data])

  const history = useAsync<CalculationSummary[]>(() => (female ? api.listCalculations(female.id) : Promise.resolve([])))
  const [lastFemale, setLastFemale] = useState<string | undefined>(undefined)
  if (lastFemale !== female?.id) { setLastFemale(female?.id); history.reload() }

  const canRun = !!female && bulls.size > 0 && !!profile && !running

  async function run() {
    if (!female || !profile) return
    setRunning(true); setError(null)
    try {
      const r = await api.runMating({ femaleId: female.id, bullIds: [...bulls], profile, save: true })
      setResponse(r)
      if (r.saved) { toast('success', 'Cálculo salvo no histórico.'); history.reload() }
    } catch (e) {
      setError(errorMessage(e)); setResponse(null)
    } finally {
      setRunning(false)
    }
  }

  if (profiles.error) return <div role="alert" className="notice notice-error">{profiles.error}</div>

  return (
    <>
      <PageHeader title="Acasalamentos" subtitle="Ranking transparente de touros para uma matriz. Cada nota pode ser aberta em “Ver cálculo”." />

      <section className="card block" aria-labelledby="step-female">
        <h2 id="step-female">1. Matriz</h2>
        <FemalePicker selected={female} onSelect={(a) => { setFemale(a); setResponse(null) }} />
      </section>

      <section className="card block" aria-labelledby="step-bulls">
        <h2 id="step-bulls">2. Touros candidatos</h2>
        <CandidatePicker selected={bulls} onChange={setBulls} />
      </section>

      <section className="card block" aria-labelledby="step-profile">
        <h2 id="step-profile">3. Objetivo e pesos</h2>
        {profiles.data && profile
          ? <ProfileEditor profiles={profiles.data} profile={profile} onChange={setProfile} onProfilesChanged={profiles.reload} />
          : <p role="status" className="muted">Carregando perfis…</p>}
      </section>

      <div className="form-actions sticky-actions">
        <button className="btn btn-primary" onClick={run} disabled={!canRun}>{running ? 'Calculando…' : 'Calcular ranking'}</button>
      </div>
      {!female && <p className="muted small">Escolha uma matriz para calcular.</p>}
      {error && <div role="alert" className="notice notice-error">{error}</div>}

      {response && (
        <section className="spaced" aria-label="Resultado do cálculo">
          <ResultsView output={response.output} saved={response.saved} onViewCalculation={(bullId) => setView({ bullId })} />
        </section>
      )}

      {female && history.data && history.data.length > 0 && (
        <section className="card block" aria-label="Cálculos anteriores">
          <h2>Cálculos anteriores de {animalTitle(female)}</h2>
          <div className="table-wrap">
            <table className="data-table compact">
              <thead><tr><th>Quando</th><th>Touro</th><th>Perfil</th><th className="right">Score</th><th>Motor</th><th><span className="sr-only">Ações</span></th></tr></thead>
              <tbody>
                {history.data.map((h) => (
                  <tr key={h.id}>
                    <td>{formatDateTime(h.createdAt)}</td><td>{h.bullName}{h.isDemo && <span className="badge badge-demo"> DEMO</span>}</td><td>{h.profileName}</td>
                    <td className="right">{h.score == null ? <span className="muted">não calculável</span> : <strong>{fmt2(h.score)}</strong>}</td><td className="mono">{h.engineVersion}</td>
                    <td><button className="btn btn-link" onClick={() => setView({ matingId: h.id })}>Ver cálculo</button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      )}
      {!response && !female && <EmptyState title="Comece escolhendo uma matriz">O ranking só considera dados e métodos que o aplicativo realmente suporta.</EmptyState>}

      {view && (
        <CalculationDrawer
          matingId={view.matingId}
          snapshot={view.bullId && response ? response.snapshots[view.bullId] : undefined}
          onClose={() => setView(null)}
        />
      )}
    </>
  )
}
