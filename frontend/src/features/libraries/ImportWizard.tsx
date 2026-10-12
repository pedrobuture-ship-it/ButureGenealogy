import { useMemo, useState } from 'react'
import { Field } from '../../components/Field'
import { useToast } from '../../components/Toast'
import { useAsync } from '../../hooks/useAsync'
import type { ImportInspection, ImportRequest, ImportValidation, InstallResult } from '../../types/domain'
import { api, errorMessage, fieldErrors } from '../../utils/api'
import { traitCategoryLabel, valueKindLabel } from '../../utils/labels'
import { LicenseBadge } from './LicenseBadge'
import { ValidationReport } from './ValidationReport'
import {
  buildRequest, emptyMeta, slugify, suggestCode, type ColumnMapping, type MetaForm,
} from './importRequest'

type Step = 'file' | 'details' | 'confirm' | 'done'

const GENETIC_KINDS = ['DEP', 'EPD', 'EBV', 'PTA', 'INDEX']

export function ImportWizard({ onInstalled }: { onInstalled: () => void }) {
  const toast = useToast()
  const [step, setStep] = useState<Step>('file')
  const [pathInput, setPathInput] = useState('')
  const [inspection, setInspection] = useState<ImportInspection | null>(null)
  const [meta, setMeta] = useState<MetaForm>(emptyMeta)
  const [mappings, setMappings] = useState<ColumnMapping[]>([])
  const [validation, setValidation] = useState<ImportValidation | null>(null)
  const [policy, setPolicy] = useState<'skip' | 'create'>('skip')
  const [farmId, setFarmId] = useState('')
  const [accept, setAccept] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [fieldErrs, setFieldErrs] = useState<Record<string, string>>({})
  const [result, setResult] = useState<InstallResult | null>(null)

  async function inspect(path: string) {
    if (!path.trim()) return
    setBusy(true); setError(null)
    try {
      const r = await api.inspectImport(path.trim())
      setInspection(r)
      setPathInput(r.path)
      if (r.kind === 'genlib') {
        setValidation(r.validation)
        setStep('confirm')
      } else {
        setMeta({ ...emptyMeta, idColumn: r.suggestedIdColumn, nameColumn: r.suggestedNameColumn, decimalComma: r.suggestedDecimalComma })
        setMappings(r.columns.map((c) => ({
          column: c, enabled: false, code: suggestCode(c), name: c, evaluationType: 'DEP', unit: '', category: 'other',
          accuracyColumn: '', percentileColumn: '', existingTraitId: '', percentileSemantics: '', popMean: '', popSd: '',
        })))
        setStep('details')
      }
    } catch (e) {
      setError(errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  async function pick() {
    setError(null)
    try {
      const p = await api.pickLibraryFile()
      if (p) await inspect(p)
    } catch (e) {
      setError(errorMessage(e))
    }
  }

  const baseRequest = (): ImportRequest => {
    if (inspection?.kind === 'genlib') {
      return { path: inspection.path, manifest: inspection.manifest!, traits: inspection.traits, licenseText: '', unmatchedPolicy: policy, farmId, acceptResponsibility: accept }
    }
    return { ...buildRequest(inspection!.path, meta, mappings), unmatchedPolicy: policy, farmId, acceptResponsibility: accept }
  }

  async function validate() {
    setBusy(true); setError(null); setFieldErrs({})
    try {
      setValidation(await api.validateImport(baseRequest()))
      setStep('confirm')
    } catch (e) {
      setError(errorMessage(e)); setFieldErrs(fieldErrors(e))
    } finally {
      setBusy(false)
    }
  }

  async function install() {
    setBusy(true); setError(null); setFieldErrs({})
    try {
      const r = await api.installLibrary(baseRequest())
      setResult(r)
      setStep('done')
      toast('success', 'Biblioteca instalada.')
    } catch (e) {
      setError(errorMessage(e)); setFieldErrs(fieldErrors(e))
    } finally {
      setBusy(false)
    }
  }

  function reset() {
    setStep('file'); setInspection(null); setValidation(null); setPathInput(''); setMeta(emptyMeta); setMappings([])
    setAccept(false); setPolicy('skip'); setFarmId(''); setError(null); setResult(null)
  }

  return (
    <div className="wizard">
      <ol className="steps" aria-label="Etapas da importação">
        {[['file', '1. Arquivo'], ['details', '2. Dados e mapeamento'], ['confirm', '3. Validação'], ['done', '4. Concluído']].map(([id, label]) => (
          <li key={id} aria-current={step === id ? 'step' : undefined} className={step === id ? 'step active' : 'step'}>{label}</li>
        ))}
      </ol>

      {error && <div role="alert" className="notice notice-error">{error}</div>}

      {step === 'file' && (
        <div className="card">
          <h2>Escolha o arquivo</h2>
          <p className="muted">Aceita pacotes <strong>.genlib</strong> e arquivos <strong>CSV</strong> ou <strong>JSON</strong> (com mapeamento de colunas).</p>
          <div className="row-gap">
            <button className="btn btn-primary" onClick={pick} disabled={busy}>Escolher arquivo…</button>
            <span className="muted">ou informe o caminho:</span>
            <input aria-label="Caminho do arquivo" className="grow" value={pathInput} onChange={(e) => setPathInput(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Enter') void inspect(pathInput) }} placeholder="/caminho/para/arquivo.genlib" />
            <button className="btn" onClick={() => inspect(pathInput)} disabled={busy || !pathInput.trim()}>Analisar</button>
          </div>
          <p className="notice notice-info">{NOTICE}</p>
        </div>
      )}

      {step === 'details' && inspection && (
        <DetailsStep inspection={inspection} meta={meta} setMeta={setMeta} mappings={mappings} setMappings={setMappings}
          errors={fieldErrs} busy={busy} onBack={reset} onNext={validate} />
      )}

      {step === 'confirm' && inspection && validation && (
        <ConfirmStep inspection={inspection} validation={validation} policy={policy} setPolicy={setPolicy} farmId={farmId} setFarmId={setFarmId}
          accept={accept} setAccept={setAccept} busy={busy} errors={fieldErrs}
          onBack={() => (inspection.kind === 'genlib' ? reset() : setStep('details'))} onInstall={install} />
      )}

      {step === 'done' && result && (
        <div className="card" role="status">
          <h2>Biblioteca instalada</h2>
          <p><strong>{result.library.name}</strong> v{result.library.version}</p>
          <ul className="plain-list">
            <li>{result.evaluationsInserted.toLocaleString('pt-BR')} avaliações importadas</li>
            <li>{result.animalsCreated.toLocaleString('pt-BR')} animais cadastrados</li>
            <li>{result.recordsSkipped.toLocaleString('pt-BR')} registros ignorados</li>
          </ul>
          <div className="row-gap">
            <button className="btn btn-primary" onClick={onInstalled}>Ver bibliotecas instaladas</button>
            <button className="btn" onClick={reset}>Importar outra</button>
          </div>
        </div>
      )}
    </div>
  )
}

const NOTICE = 'A responsabilidade pela autorização de uso e distribuição de dados importados manualmente é do usuário.'

// ---------------------------------------------------------------- step 2

interface DetailsProps {
  inspection: ImportInspection
  meta: MetaForm
  setMeta: (m: MetaForm) => void
  mappings: ColumnMapping[]
  setMappings: (m: ColumnMapping[]) => void
  errors: Record<string, string>
  busy: boolean
  onBack: () => void
  onNext: () => void
}

function DetailsStep({ inspection, meta, setMeta, mappings, setMappings, errors, busy, onBack, onNext }: DetailsProps) {
  const set = <K extends keyof MetaForm>(k: K, v: MetaForm[K]) => setMeta({ ...meta, [k]: v })
  const err = (k: string) => errors[`manifest.${k}`]
  const reserved = new Set([meta.idColumn, meta.nameColumn, meta.sexColumn].filter(Boolean))
  const mappable = mappings.filter((m) => !reserved.has(m.column))
  const catalogue = useAsync(() => api.listTraits({}))
  const genetic = (catalogue.data ?? []).filter((t) => GENETIC_KINDS.includes(t.evaluationType) && t.sourceDataset === '')
  const colOptions = inspection.columns
  const canContinue = meta.name.trim() && meta.version.trim() && meta.program.trim() && meta.idColumn && mappings.some((m) => m.enabled)

  const updateMapping = (col: string, patch: Partial<ColumnMapping>) =>
    setMappings(mappings.map((m) => (m.column === col ? { ...m, ...patch } : m)))

  function applyCatalogueTrait(col: string, traitId: string) {
    const t = genetic.find((x) => x.id === traitId)
    if (!t) { updateMapping(col, { existingTraitId: '' }); return }
    updateMapping(col, { existingTraitId: t.id, code: t.code, name: t.name, unit: t.unit, category: t.category, evaluationType: t.evaluationType, enabled: true })
  }

  const togglePurpose = (p: string) =>
    set('purposes', meta.purposes.includes(p) ? meta.purposes.filter((x) => x !== p) : [...meta.purposes, p])

  return (
    <>
      <section className="card block">
        <h2>Sobre este conjunto de dados</h2>
        <p className="muted">{inspection.rowCount.toLocaleString('pt-BR')} linhas · {inspection.columns.length} colunas</p>
        <div className="form-grid">
          <Field label="Nome" error={err('name')}>{(id, d) => <input id={id} aria-describedby={d} value={meta.name} onChange={(e) => setMeta({ ...meta, name: e.target.value, id: slugify(e.target.value) })} />}</Field>
          <Field label="Identificador" error={err('id')} hint="Minúsculas, números, ponto e hífen.">{(id, d) => <input id={id} aria-describedby={d} value={meta.id} onChange={(e) => set('id', e.target.value)} />}</Field>
          <Field label="Versão" error={err('version')}>{(id, d) => <input id={id} aria-describedby={d} value={meta.version} onChange={(e) => set('version', e.target.value)} placeholder="2026.1" />}</Field>
          <Field label="Programa de avaliação" error={err('evaluation.program')} hint="Origem das avaliações (obrigatório para a proveniência).">{(id, d) => <input id={id} aria-describedby={d} value={meta.program} onChange={(e) => set('program', e.target.value)} />}</Field>
          <Field label="Release" optional hint="Se vazio, usa a versão.">{(id) => <input id={id} value={meta.release} onChange={(e) => set('release', e.target.value)} />}</Field>
          <Field label="Base genética" optional>{(id) => <input id={id} value={meta.geneticBase} onChange={(e) => set('geneticBase', e.target.value)} />}</Field>
          <Field label="Espécie" error={err('species')}>{(id, d) => <input id={id} aria-describedby={d} value={meta.species} onChange={(e) => set('species', e.target.value)} />}</Field>
          <Field label="Raças (códigos separados por vírgula)" error={err('breeds')}>{(id, d) => <input id={id} aria-describedby={d} value={meta.breeds} onChange={(e) => set('breeds', e.target.value)} placeholder="NEL, GIR" />}</Field>
          <Field label="Organização" optional>{(id) => <input id={id} value={meta.organization} onChange={(e) => set('organization', e.target.value)} />}</Field>
          <Field label="URL da fonte" optional error={err('sourceUrl')}>{(id, d) => <input id={id} aria-describedby={d} value={meta.sourceUrl} onChange={(e) => set('sourceUrl', e.target.value)} />}</Field>
          <Field label="Data de lançamento" optional error={err('releaseDate')}>{(id, d) => <input id={id} aria-describedby={d} type="date" value={meta.releaseDate} onChange={(e) => set('releaseDate', e.target.value)} />}</Field>
          <fieldset className="field">
            <legend>Finalidade</legend>
            <div className="row-gap">
              {(['beef', 'dairy', 'dual'] as const).map((p) => (
                <label key={p} className="check"><input type="checkbox" checked={meta.purposes.includes(p)} onChange={() => togglePurpose(p)} /> {p === 'beef' ? 'Corte' : p === 'dairy' ? 'Leite' : 'Dupla aptidão'}</label>
              ))}
            </div>
            {err('purpose') && <p className="field-error" role="alert">{err('purpose')}</p>}
          </fieldset>
        </div>
      </section>

      <section className="card block">
        <h2>Licença</h2>
        <p className="muted">Deixe em branco se você não souber: a licença ficará marcada como <strong>desconhecida</strong>. “Disponível gratuitamente” não é uma licença.</p>
        <div className="form-grid">
          <Field label="Nome da licença" optional>{(id) => <input id={id} value={meta.licenseName} onChange={(e) => set('licenseName', e.target.value)} placeholder="ex.: CC-BY-4.0" />}</Field>
          <Field label="URL da licença" optional error={err('license.url')}>{(id, d) => <input id={id} aria-describedby={d} value={meta.licenseUrl} onChange={(e) => set('licenseUrl', e.target.value)} />}</Field>
          <label className="check span-2"><input type="checkbox" checked={meta.redistributionAllowed} disabled={!meta.licenseName.trim()} onChange={(e) => set('redistributionAllowed', e.target.checked)} /> A licença permite redistribuição</label>
          <div className="span-2"><Field label="Texto da licença" optional>{(id) => <textarea id={id} rows={3} value={meta.licenseText} onChange={(e) => set('licenseText', e.target.value)} />}</Field></div>
        </div>
      </section>

      <section className="card block">
        <h2>Leitura do arquivo</h2>
        <div className="form-grid">
          <Field label="Coluna de identificação (registro)" error={err('data.idColumn')}>{(id, d) => (
            <select id={id} aria-describedby={d} value={meta.idColumn} onChange={(e) => set('idColumn', e.target.value)}>
              <option value="">Selecione…</option>{colOptions.map((c) => <option key={c} value={c}>{c}</option>)}
            </select>)}
          </Field>
          <Field label="Coluna do nome" optional>{(id) => (
            <select id={id} value={meta.nameColumn} onChange={(e) => set('nameColumn', e.target.value)}>
              <option value="">Nenhuma</option>{colOptions.map((c) => <option key={c} value={c}>{c}</option>)}
            </select>)}
          </Field>
          <Field label="Coluna do sexo" optional>{(id) => (
            <select id={id} value={meta.sexColumn} onChange={(e) => set('sexColumn', e.target.value)}>
              <option value="">Nenhuma (usar o sexo padrão)</option>{colOptions.map((c) => <option key={c} value={c}>{c}</option>)}
            </select>)}
          </Field>
          <Field label="Sexo padrão" hint="Usado quando não há coluna de sexo.">{(id) => (
            <select id={id} value={meta.defaultSex} disabled={!!meta.sexColumn} onChange={(e) => set('defaultSex', e.target.value as 'male' | 'female')}>
              <option value="male">Macho (catálogo de touros)</option><option value="female">Fêmea</option>
            </select>)}
          </Field>
          <label className="check"><input type="checkbox" checked={meta.decimalComma} onChange={(e) => set('decimalComma', e.target.checked)} /> Números usam vírgula decimal (1.234,5)</label>
          <Field label="Escala da acurácia">{(id) => (
            <select id={id} value={meta.accuracyScale} onChange={(e) => set('accuracyScale', e.target.value as 'fraction' | 'percent')}>
              <option value="fraction">Fração (0 a 1)</option><option value="percent">Percentual (0 a 100)</option>
            </select>)}
          </Field>
        </div>
        {inspection.sampleRows.length > 0 && (
          <div className="table-wrap sample">
            <table className="data-table compact">
              <caption className="sr-only">Primeiras linhas do arquivo</caption>
              <thead><tr>{inspection.columns.map((c) => <th key={c}>{c}</th>)}</tr></thead>
              <tbody>{inspection.sampleRows.map((r, i) => <tr key={i}>{r.map((v, j) => <td key={j}>{v}</td>)}</tr>)}</tbody>
            </table>
          </div>
        )}
      </section>

      <section className="card block">
        <h2>Mapeamento das colunas</h2>
        <p className="muted">Marque as colunas que contêm valores de avaliação e diga a que característica e tipo cada uma corresponde. O tipo (DEP, EPD, EBV, PTA, índice) nunca é deduzido: você o informa.</p>
        <p className="muted small">Se informar um percentil, diga o que ele significa: o motor <strong>não usa</strong> percentis sem essa declaração (nunca presume o sentido). Médias e desvios populacionais, quando o dataset os fornece, habilitam a normalização por escore z.</p>
        {errors.traits && <p role="alert" className="field-error">{errors.traits}</p>}
        <div className="table-wrap">
          <table className="data-table compact mapping">
            <caption className="sr-only">Mapeamento de colunas para características</caption>
            <thead><tr><th>Usar</th><th>Coluna</th><th>Característica (catálogo)</th><th>Código</th><th>Nome</th><th>Tipo</th><th>Unidade</th><th>Categoria</th><th>Acurácia</th><th>Percentil</th></tr></thead>
            <tbody>
              {mappable.map((m) => (
                <tr key={m.column} className={m.enabled ? '' : 'row-inactive'}>
                  <td><input type="checkbox" aria-label={`Usar a coluna ${m.column}`} checked={m.enabled} onChange={(e) => updateMapping(m.column, { enabled: e.target.checked })} /></td>
                  <td className="mono">{m.column}</td>
                  <td>
                    <select aria-label={`Característica do catálogo para ${m.column}`} value={m.existingTraitId} disabled={!m.enabled} onChange={(e) => applyCatalogueTrait(m.column, e.target.value)}>
                      <option value="">Nova característica</option>
                      {genetic.map((t) => <option key={t.id} value={t.id}>{t.name} ({t.evaluationType})</option>)}
                    </select>
                  </td>
                  <td><input aria-label={`Código de ${m.column}`} value={m.code} disabled={!m.enabled || !!m.existingTraitId} onChange={(e) => updateMapping(m.column, { code: e.target.value })} /></td>
                  <td><input aria-label={`Nome de ${m.column}`} value={m.name} disabled={!m.enabled || !!m.existingTraitId} onChange={(e) => updateMapping(m.column, { name: e.target.value })} /></td>
                  <td>
                    <select aria-label={`Tipo de ${m.column}`} value={m.evaluationType} disabled={!m.enabled || !!m.existingTraitId} onChange={(e) => updateMapping(m.column, { evaluationType: e.target.value })}>
                      {GENETIC_KINDS.map((k) => <option key={k} value={k}>{valueKindLabel[k]}</option>)}
                    </select>
                  </td>
                  <td><input aria-label={`Unidade de ${m.column}`} className="narrow" value={m.unit} disabled={!m.enabled || !!m.existingTraitId} onChange={(e) => updateMapping(m.column, { unit: e.target.value })} /></td>
                  <td>
                    <select aria-label={`Categoria de ${m.column}`} value={m.category} disabled={!m.enabled || !!m.existingTraitId} onChange={(e) => updateMapping(m.column, { category: e.target.value })}>
                      {Object.entries(traitCategoryLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
                    </select>
                  </td>
                  <td>
                    <select aria-label={`Coluna de acurácia de ${m.column}`} value={m.accuracyColumn} disabled={!m.enabled} onChange={(e) => updateMapping(m.column, { accuracyColumn: e.target.value })}>
                      <option value="">Sem acurácia</option>{colOptions.filter((c) => c !== m.column).map((c) => <option key={c} value={c}>{c}</option>)}
                    </select>
                  </td>
                  <td>
                    <select aria-label={`Coluna de percentil de ${m.column}`} value={m.percentileColumn} disabled={!m.enabled} onChange={(e) => updateMapping(m.column, { percentileColumn: e.target.value })}>
                      <option value="">Sem percentil</option>{colOptions.filter((c) => c !== m.column).map((c) => <option key={c} value={c}>{c}</option>)}
                    </select>
                    {m.enabled && m.percentileColumn && (
                      <select aria-label={`Significado do percentil de ${m.column}`} value={m.percentileSemantics} onChange={(e) => updateMapping(m.column, { percentileSemantics: e.target.value as ColumnMapping['percentileSemantics'] })}>
                        <option value="">Significado não informado</option>
                        <option value="ascending">Valor maior ⇒ percentil maior</option>
                        <option value="descending">Valor maior ⇒ percentil menor (ex.: “top 10%”)</option>
                      </select>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <div className="form-actions sticky-actions">
        <button className="btn" onClick={onBack} disabled={busy}>Voltar</button>
        <button className="btn btn-primary" onClick={onNext} disabled={busy || !canContinue}>{busy ? 'Validando…' : 'Validar importação'}</button>
      </div>
    </>
  )
}

// ---------------------------------------------------------------- step 3

interface ConfirmProps {
  inspection: ImportInspection
  validation: ImportValidation
  policy: 'skip' | 'create'
  setPolicy: (p: 'skip' | 'create') => void
  farmId: string
  setFarmId: (id: string) => void
  accept: boolean
  setAccept: (b: boolean) => void
  busy: boolean
  errors: Record<string, string>
  onBack: () => void
  onInstall: () => void
}

function ConfirmStep({ inspection, validation, policy, setPolicy, farmId, setFarmId, accept, setAccept, busy, errors, onBack, onInstall }: ConfirmProps) {
  const farms = useAsync(() => api.listFarms())
  const m = validation.manifest
  const needsFarm = policy === 'create'
  const ready = validation.canInstall && accept && (!needsFarm || farmId)
  const memo = useMemo(() => farms.data ?? [], [farms.data])

  return (
    <>
      <section className="card block">
        <h2>{m.name} <span className="muted">v{m.version}</span></h2>
        <p>
          <LicenseBadge status={validation.licenseStatus} name={m.license.name} />{' '}
          {m.license.name && <span className="muted">{m.license.name}</span>}
          {validation.duplicate && <span className="badge badge-warn"> já instalada</span>}
        </p>
        <p className="muted small">Arquivo: <span className="mono">{inspection.path}</span> · Programa: {m.evaluation.program} · Origem: {inspection.kind.toUpperCase()}</p>
      </section>

      <ValidationReport report={validation.report} />

      <section className="card block">
        <h2>Animais não cadastrados</h2>
        <fieldset className="field">
          <legend className="sr-only">O que fazer com registros que não correspondem a animais do rebanho</legend>
          <label className="check"><input type="radio" name="policy" checked={policy === 'skip'} onChange={() => setPolicy('skip')} /> Ignorar (importar apenas animais que já estão no rebanho)</label>
          <label className="check"><input type="radio" name="policy" checked={policy === 'create'} onChange={() => setPolicy('create')} /> Cadastrar como novos animais (útil para catálogos de touros)</label>
        </fieldset>
        {needsFarm && (
          <Field label="Propriedade para os novos animais" error={errors.farmId}>{(id, d) => (
            <select id={id} aria-describedby={d} value={farmId} onChange={(e) => setFarmId(e.target.value)}>
              <option value="">Selecione…</option>{memo.map((f) => <option key={f.id} value={f.id}>{f.name}</option>)}
            </select>)}
          </Field>
        )}
      </section>

      <section className="card block">
        <p className="notice notice-info">{validation.notice}</p>
        <label className="check"><input type="checkbox" checked={accept} onChange={(e) => setAccept(e.target.checked)} /> Declaro que tenho autorização para usar estes dados.</label>
        {errors.acceptResponsibility && <p role="alert" className="field-error">{errors.acceptResponsibility}</p>}
      </section>

      <div className="form-actions sticky-actions">
        <button className="btn" onClick={onBack} disabled={busy}>Voltar</button>
        <button className="btn btn-primary" onClick={onInstall} disabled={busy || !ready}>{busy ? 'Instalando…' : 'Instalar biblioteca'}</button>
      </div>
    </>
  )
}
