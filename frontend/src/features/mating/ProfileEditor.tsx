import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Field } from '../../components/Field'
import { useToast } from '../../components/Toast'
import { useAsync } from '../../hooks/useAsync'
import type { Profile, TraitWeight } from '../../types/engine'
import { api, errorMessage, fieldErrors } from '../../utils/api'
import { directionLabel, valueKindLabel } from '../../utils/labels'
import { fmtPercent } from '../../utils/format'
import { docPath, docSection } from './docLinks'
import { catalogKey, cloneProfile, fractionToPercentText, newTraitWeight, nominalShares, numberOrNull, percentToFraction, traitKey } from './profileUtils'

interface Props {
  profiles: Profile[]
  profile: Profile
  onChange: (p: Profile) => void
  onProfilesChanged: () => void
}

const GENETIC = ['DEP', 'EPD', 'EBV', 'PTA', 'INDEX']

export function ProfileEditor({ profiles, profile, onChange, onProfilesChanged }: Props) {
  const toast = useToast()
  const catalogue = useAsync(() => api.listTraits({}))
  const [adding, setAdding] = useState('')
  const [saveName, setSaveName] = useState('')
  const [saveError, setSaveError] = useState<string | null>(null)

  const traitInfo = useMemo(() => {
    const m = new Map<string, { name: string; direction: string; unit: string }>()
    for (const t of catalogue.data ?? []) if (!m.has(catalogKey(t))) m.set(catalogKey(t), { name: t.name, direction: t.optimizationDirection, unit: t.unit })
    return m
  }, [catalogue.data])
  const shares = nominalShares(profile.traits)
  const set = (patch: Partial<Profile>) => onChange({ ...profile, ...patch })
  const setTrait = (i: number, patch: Partial<TraitWeight>) => set({ traits: profile.traits.map((t, j) => (j === i ? { ...t, ...patch } : t)) })
  const geneticTraits = (catalogue.data ?? []).filter((t) => GENETIC.includes(t.evaluationType))
  const available = geneticTraits.filter((t, i, a) => a.findIndex((x) => catalogKey(x) === catalogKey(t)) === i && !profile.traits.some((p) => traitKey(p) === catalogKey(t)))
  const isSaved = !profile.id.startsWith('builtin:') && profile.id !== ''

  function addTrait() {
    const [code, type] = adding.split('|')
    if (!code) return
    set({ traits: [...profile.traits, newTraitWeight(code, type)] })
    setAdding('')
  }

  async function save(asNew: boolean) {
    setSaveError(null)
    try {
      const p = await api.saveMatingProfile({ ...profile, id: asNew ? '' : profile.id, name: asNew ? saveName : profile.name })
      toast('success', 'Perfil salvo.')
      setSaveName('')
      onProfilesChanged()
      onChange(p)
    } catch (e) {
      setSaveError(fieldErrors(e).name ?? errorMessage(e))
    }
  }

  async function remove() {
    try {
      await api.deleteMatingProfile(profile.id)
      toast('success', 'Perfil excluído.')
      onProfilesChanged()
      onChange(cloneProfile(profiles[0]))
    } catch (e) {
      toast('error', errorMessage(e))
    }
  }

  const f = profile.filters
  return (
    <div>
      <div className="row-gap">
        <Field label="Perfil de objetivo">{(id) => (
          <select id={id} value={profile.id} onChange={(e) => { const p = profiles.find((x) => x.id === e.target.value); if (p) onChange(cloneProfile(p)) }}>
            {profiles.map((p) => <option key={p.id} value={p.id}>{p.name}{p.isDemo ? ' (DEMO)' : ''}</option>)}
          </select>)}
        </Field>
      </div>
      {profile.isDemo && (
        <div className="notice notice-warn" role="note"><strong>{profile.demoNotice}</strong><span>Pesos e direções deste perfil são arbitrários, apenas para demonstrar o motor.</span></div>
      )}
      {profile.description && !profile.isDemo && <p className="muted">{profile.description}</p>}

      <h3 className="sub-h">Características e pesos</h3>
      <p className="muted small">Os pesos são relativos: o motor os normaliza entre os componentes que puderem ser calculados. A direção só é necessária quando a característica não a define. <Link to={docPath(docSection.weights)}>Como funciona</Link></p>
      {profile.traits.length === 0 ? (
        <p className="muted">Nenhuma característica no perfil. Adicione abaixo.</p>
      ) : (
        <div className="table-wrap">
          <table className="data-table compact weights">
            <caption className="sr-only">Pesos do perfil</caption>
            <thead><tr><th>Característica</th><th>Tipo</th><th>Peso</th><th>Peso relativo</th><th>Direção</th><th>Obrigatória</th><th><span className="sr-only">Remover</span></th></tr></thead>
            <tbody>
              {profile.traits.map((t, i) => {
                const info = traitInfo.get(traitKey(t))
                const needsDirection = !t.direction && (info?.direction ?? 'INFORMATION_ONLY') === 'INFORMATION_ONLY'
                return (
                  <tr key={traitKey(t)}>
                    <td><strong>{info?.name ?? t.traitCode}</strong> <span className="muted mono">{t.traitCode}</span></td>
                    <td>{valueKindLabel[t.evaluationType] ?? t.evaluationType}</td>
                    <td><input type="number" min={0} step="any" className="narrow" aria-label={`Peso de ${t.traitCode}`} value={t.weight} onChange={(e) => setTrait(i, { weight: Number(e.target.value) })} /></td>
                    <td>{fmtPercent(shares[i], 1)}</td>
                    <td>
                      <select aria-label={`Direção de ${t.traitCode}`} value={t.direction} onChange={(e) => setTrait(i, { direction: e.target.value as TraitWeight['direction'] })}>
                        <option value="">Do catálogo ({directionLabel[info?.direction ?? 'INFORMATION_ONLY']})</option>
                        <option value="MAXIMIZE">{directionLabel.MAXIMIZE}</option><option value="MINIMIZE">{directionLabel.MINIMIZE}</option><option value="TARGET_RANGE">{directionLabel.TARGET_RANGE}</option>
                      </select>
                      {needsDirection && <p className="field-error small">Sem direção definida: escolha uma, ou a característica ficará não calculável.</p>}
                      {t.direction === 'TARGET_RANGE' && (
                        <div className="row-gap range-inputs">
                          <input aria-label={`Mínimo ideal de ${t.traitCode}`} placeholder="mín." value={t.idealMin ?? ''} onChange={(e) => setTrait(i, { idealMin: numberOrNull(e.target.value) })} />
                          <input aria-label={`Máximo ideal de ${t.traitCode}`} placeholder="máx." value={t.idealMax ?? ''} onChange={(e) => setTrait(i, { idealMax: numberOrNull(e.target.value) })} />
                          <input aria-label={`Tolerância de ${t.traitCode}`} placeholder="tolerância" value={t.targetTolerance ?? ''} onChange={(e) => setTrait(i, { targetTolerance: numberOrNull(e.target.value) })} />
                        </div>
                      )}
                    </td>
                    <td><input type="checkbox" aria-label={`${t.traitCode} obrigatória`} checked={t.required} onChange={(e) => setTrait(i, { required: e.target.checked })} /></td>
                    <td><button className="btn btn-link danger-link" onClick={() => set({ traits: profile.traits.filter((_, j) => j !== i) })}>Remover</button></td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
      <div className="row-gap spaced">
        <select aria-label="Adicionar característica" value={adding} onChange={(e) => setAdding(e.target.value)}>
          <option value="">Adicionar característica…</option>
          {available.map((t) => <option key={catalogKey(t)} value={catalogKey(t)}>{t.name} ({valueKindLabel[t.evaluationType]})</option>)}
        </select>
        <button className="btn" onClick={addTrait} disabled={!adding}>Adicionar</button>
      </div>

      <h3 className="sub-h">Parâmetros do perfil</h3>
      <div className="form-grid params">
        <Field label="Normalização" hint="Percentil só é usado se o dataset declarar o seu significado.">{(id) => (
          <select id={id} value={profile.normalization} onChange={(e) => set({ normalization: e.target.value as Profile['normalization'] })}>
            <option value="auto">Automática (percentil, senão escore z)</option><option value="percentile">Somente percentil</option><option value="zscore">Somente escore z</option>
          </select>)}
        </Field>
        <Field label="Cobertura mínima do peso (%)" hint="Abaixo disso o touro fica sem score (não calculável).">{(id) => (
          <input id={id} inputMode="decimal" value={fractionToPercentText(profile.minCoverage)} onChange={(e) => set({ minCoverage: percentToFraction(e.target.value) ?? 1 })} />)}
        </Field>
        <Field label="Bônus máximo de complementaridade (pontos)" hint="0 desativa. Ver Cálculos › Complementaridade.">{(id) => (
          <input id={id} inputMode="decimal" value={profile.complementarity.maxBonusPoints} onChange={(e) => set({ complementarity: { maxBonusPoints: numberOrNull(e.target.value) ?? 0 } })} />)}
        </Field>
        <Field label="Penalidade por baixa acurácia: limiar (%)" hint="0 desativa.">{(id) => (
          <input id={id} inputMode="decimal" value={fractionToPercentText(profile.penalties.lowAccuracy.threshold)} onChange={(e) => set({ penalties: { lowAccuracy: { ...profile.penalties.lowAccuracy, threshold: percentToFraction(e.target.value) ?? 0 } } })} />)}
        </Field>
        <Field label="Penalidade por baixa acurácia: máximo (pontos)">{(id) => (
          <input id={id} inputMode="decimal" value={profile.penalties.lowAccuracy.maxPoints} onChange={(e) => set({ penalties: { lowAccuracy: { ...profile.penalties.lowAccuracy, maxPoints: numberOrNull(e.target.value) ?? 0 } } })} />)}
        </Field>
      </div>

      <h3 className="sub-h">Filtros</h3>
      <div className="form-grid params">
        <label className="check"><input type="checkbox" checked={f.requireStock} onChange={(e) => set({ filters: { ...f, requireStock: e.target.checked } })} /> Exigir sêmen em estoque</label>
        <Field label="Mínimo de doses">{(id) => <input id={id} inputMode="numeric" value={f.minDoses} onChange={(e) => set({ filters: { ...f, minDoses: Math.max(0, Math.trunc(numberOrNull(e.target.value) ?? 0)) } })} />}</Field>
        <Field label="Custo máximo por dose (R$)" optional>{(id) => <input id={id} inputMode="decimal" value={f.maxUnitCost ?? ''} onChange={(e) => set({ filters: { ...f, maxUnitCost: numberOrNull(e.target.value) } })} />}</Field>
        <Field label="Acurácia mínima (%)" optional hint="Se definida, acurácia desconhecida também exclui.">{(id) => <input id={id} inputMode="decimal" value={fractionToPercentText(f.minAccuracy)} onChange={(e) => set({ filters: { ...f, minAccuracy: percentToFraction(e.target.value) } })} />}</Field>
        <label className="check span-2"><input type="checkbox" checked={f.excludeDirectAncestors} onChange={(e) => set({ filters: { ...f, excludeDirectAncestors: e.target.checked } })} /> Excluir touros que são ancestrais da matriz</label>
      </div>

      <h3 className="sub-h">Salvar perfil</h3>
      <div className="row-gap">
        <input aria-label="Nome do novo perfil" placeholder="Nome do perfil" value={saveName} onChange={(e) => setSaveName(e.target.value)} />
        <button className="btn" onClick={() => save(true)} disabled={!saveName.trim()}>Salvar como novo</button>
        {isSaved && <button className="btn" onClick={() => save(false)}>Atualizar “{profile.name}”</button>}
        {isSaved && <button className="btn btn-danger-outline" onClick={remove}>Excluir perfil</button>}
      </div>
      {saveError && <p role="alert" className="field-error">{saveError}</p>}
    </div>
  )
}
