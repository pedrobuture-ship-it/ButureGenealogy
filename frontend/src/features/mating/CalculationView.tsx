import { useState } from 'react'
import { Link } from 'react-router-dom'
import { DemoBadge } from '../../components/Badge'
import type { ComponentResult, PairSnapshot } from '../../types/engine'
import { animalTitle, fmt2, fmtPercent, formatDate, formatNumber, signed2 } from '../../utils/format'
import { directionLabel, valueKindLabel } from '../../utils/labels'
import { directionSection, docPath, docSection } from './docLinks'

const methodLabel: Record<string, string> = { percentile: 'Percentil', zscore: 'Escore z (normal)', target_range: 'Faixa-alvo' }

/** Renders a saved snapshot. Every number shown comes from the snapshot; nothing is recalculated here. */
export function CalculationView({ snapshot: s }: { snapshot: PairSnapshot }) {
  const r = s.result
  const f = r.final
  const title = `${animalTitle(s.female)} × ${animalTitle(s.bull)}`
  const demoNotice = s.profile.demoNotice

  return (
    <div className="calc">
      {(s.isDemo || s.profile.isDemo) && (
        <div className="notice notice-warn" role="note">
          <strong>DEMO / SYNTHETIC</strong>
          <span>{demoNotice || 'Este cálculo usa dados fictícios e não representa recomendação zootécnica.'}</span>
        </div>
      )}
      <header className="calc-hero">
        <p className="muted">{title}</p>
        <h3 className="calc-title">Score final</h3>
        {f ? (
          <p className="calc-score" aria-label="Score final">{fmt2(f.final)} <span className="muted">/ 100</span></p>
        ) : (
          <>
            <p className="calc-score calc-nc" aria-label="Score final">Não calculável</p>
            <ul className="plain-list" aria-label="Motivos">{r.reasons.map((x, i) => <li key={i}>{x}</li>)}</ul>
          </>
        )}
        {f && <p className="muted small">Cobertura: {fmtPercent(f.coverage, 1)} do peso do perfil{r.tied ? ' · empate com outro candidato' : ''}{r.rank ? ` · posição ${r.rank} de ${s.candidates}` : ''}</p>}
      </header>

      <section aria-label="Componentes">
        <h3 className="calc-h">Componentes</h3>
        <p className="muted small">Peso e contribuição usam o peso normalizado entre os componentes calculáveis. <Link to={docPath(docSection.weights)}>Como os pesos funcionam</Link></p>
        {r.components.map((c) => <ComponentCard key={c.traitCode + c.evaluationType} c={c} snapshot={s} />)}
      </section>

      <section aria-label="Complementaridade">
        <h3 className="calc-h">Complementaridade <span className="calc-num">{r.complementarity.status === 'calculated' ? signed2(r.complementarity.points) : '—'}</span></h3>
        <p className="small">{r.complementarity.explanation} <Link to={docPath(docSection.complementarity)}>Saiba mais</Link></p>
        {r.complementarity.items.length > 0 && (
          <table className="data-table compact">
            <thead><tr><th>Característica</th><th className="right">Score do touro</th><th className="right">Score da matriz</th><th className="right">Ganho</th><th className="right">Pontos</th></tr></thead>
            <tbody>
              {r.complementarity.items.map((it) => (
                <tr key={it.traitCode}>
                  <td>{it.name}{it.status !== 'calculated' && <span className="badge badge-muted"> não calculável</span>}</td>
                  <td className="right">{fmt2(it.bullScore)}</td><td className="right">{fmt2(it.damScore)}</td>
                  <td className="right">{it.gain == null ? '—' : fmt2(it.gain)}</td><td className="right">{it.status === 'calculated' ? signed2(it.points) : '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
        {r.complementarity.source.parameter && <p className="muted small">Parâmetro: {r.complementarity.source.parameter}. Fórmula: {r.complementarity.formula}</p>}
        {r.complementarity.warnings.map((w, i) => <p key={i} className="small warn-line">{w}</p>)}
      </section>

      <section aria-label="Penalidades">
        <h3 className="calc-h">Penalidades <span className="calc-num">{f ? (f.penalties > 0 ? signed2(-f.penalties) : fmt2(0)) : '—'}</span></h3>
        {r.penalties.map((p) => (
          <div key={p.name} className="calc-pen">
            <strong>{p.name}</strong>{' '}
            <span className={p.status === 'calculated' ? 'badge badge-info' : 'badge badge-muted'}>{p.status === 'calculated' ? signed2(-p.points) : p.status === 'not_applied' ? 'não aplicada' : 'não calculável'}</span>
            <p className="small">{p.explanation}{p.name.includes('endogamia') && <> <Link to={docPath(docSection.relationship)}>Saiba mais</Link></>}</p>
            {p.source.parameter && <p className="muted small">Parâmetro: {p.source.parameter}</p>}
            {p.warnings.map((w, i) => <p key={i} className="small warn-line">{w}</p>)}
          </div>
        ))}
      </section>

      {f && (
        <section aria-label="Resultado" className="calc-result">
          <h3 className="calc-h">Resultado</h3>
          <dl className="calc-sum">
            <dt>Soma dos componentes</dt><dd>{fmt2(f.base)}</dd>
            <dt>Complementaridade</dt><dd>{signed2(f.complementarity)}</dd>
            <dt>Penalidades</dt><dd>{signed2(-f.penalties)}</dd>
            <dt className="calc-total">Score final</dt><dd className="calc-total">= {fmt2(f.final)}</dd>
          </dl>
          {f.clamped && <p className="small warn-line">Resultado limitado ao intervalo de 0 a 100 (valor sem limite: {fmt2(f.unclamped)}).</p>}
          <p className="muted small">Fórmula: {f.formula}</p>
        </section>
      )}

      {(r.warnings.length > 0 || r.pedigreeNotes.length > 0 || s.runWarnings.length > 0) && (
        <section aria-label="Avisos">
          <h3 className="calc-h">Avisos</h3>
          <ul className="plain-list">
            {[...r.warnings, ...s.runWarnings].filter((w, i, a) => a.indexOf(w) === i).map((w, i) => <li key={i} className="small warn-line">{w}</li>)}
          </ul>
        </section>
      )}

      <section aria-label="Dados usados neste cálculo">
        <h3 className="calc-h">Dados usados neste cálculo</h3>
        <dl className="kv">
          <dt>Dataset</dt>
          <dd>{s.datasets.length === 0 ? '—' : s.datasets.map((d) => <div key={d.id}>{d.name} {d.isDemo && <DemoBadge />}</div>)}</dd>
          <dt>Versão</dt>
          <dd>{s.datasets.length === 0 ? '—' : s.datasets.map((d) => <div key={d.id}>{d.version}</div>)}</dd>
          <dt>Motor</dt><dd>{s.engineVersion}</dd>
          <dt>Configuração</dt><dd>{s.profile.name}{s.profile.isDemo && <> <DemoBadge /></>}</dd>
          <dt>Executado em</dt><dd>{formatDate(s.createdAt)}</dd>
        </dl>
        {s.bases.length > 0 && (
          <details>
            <summary>Bases de comparação (programa, release, base genética)</summary>
            <ul className="plain-list small">
              {s.bases.map((b) => (
                <li key={b.traitCode + b.evaluationType}>
                  <strong>{b.traitCode}</strong>: {b.basis ? `${b.basis.program} · ${b.basis.release} · ${b.basis.geneticBase || 'base não informada'} · ${b.basis.breedCode} · ${b.basis.unit || 'sem unidade'}` : 'nenhuma base disponível'} — {b.reason}
                </li>
              ))}
            </ul>
            <p className="small"><Link to={docPath(docSection.datasets)}>Por que datasets diferentes não são comparáveis</Link> · <Link to={docPath(docSection.reproducibility)}>Reprodutibilidade</Link></p>
          </details>
        )}
      </section>
    </div>
  )
}

function ComponentCard({ c, snapshot }: { c: ComponentResult; snapshot: PairSnapshot }) {
  const [open, setOpen] = useState(false)
  const ok = c.status === 'calculated'
  const dam = snapshot.femaleEvals.find((e) => c.basis && e.basis.traitCode === c.traitCode && e.basis.evaluationType === c.evaluationType &&
    e.basis.program === c.basis.program && e.basis.release === c.basis.release && e.basis.geneticBase === c.basis.geneticBase)
  return (
    <article className={ok ? 'calc-comp' : 'calc-comp calc-comp-nc'}>
      <header>
        <h4>{c.name}</h4>
        {!ok && <span className="badge badge-warn">não calculável</span>}
      </header>
      {ok ? (
        <dl className="calc-metrics">
          <div><dt>Score normalizado</dt><dd>{fmt2(c.score)}</dd></div>
          <div><dt>Peso</dt><dd>{fmtPercent(c.weight, 1)}</dd></div>
          <div><dt>Contribuição</dt><dd>{fmt2(c.contribution)}</dd></div>
        </dl>
      ) : (
        <>
          <p className="small">{c.explanation}</p>
          {c.warnings.map((w, i) => <p key={i} className="small warn-line">{w}</p>)}
        </>
      )}
      {ok && (
        <>
          <button className="btn btn-link" aria-expanded={open} onClick={() => setOpen(!open)}>Como chegamos nisso?</button>
          {open && (
            <dl className="kv how">
              <dt>Valor da matriz</dt><dd>{c.damValue != null ? `${formatNumber(c.damValue)} ${c.unit}` : 'sem avaliação comparável'}{dam && dam.percentile != null && <span className="muted"> (percentil {formatNumber(dam.percentile, 0)})</span>}</dd>
              <dt>Valor do touro</dt><dd>{c.inputValue != null ? `${formatNumber(c.inputValue)} ${c.unit}` : '—'}{c.percentile != null && <span className="muted"> (percentil {formatNumber(c.percentile, 0)})</span>}{c.accuracy != null && <span className="muted"> · acurácia {fmtPercent(c.accuracy)}</span>}</dd>
              <dt>Tipo da avaliação</dt><dd>{valueKindLabel[c.evaluationType] ?? c.evaluationType}</dd>
              <dt>Dataset</dt><dd>{c.source.datasetName || '—'} {c.source.isDemo && <DemoBadge />}</dd>
              <dt>Versão</dt><dd>{c.source.datasetVersion || '—'} · programa {c.source.program}, release {c.source.release}{c.source.geneticBase ? `, base ${c.source.geneticBase}` : ''}</dd>
              <dt>Unidade</dt><dd>{c.unit || '—'}</dd>
              <dt>Direção</dt><dd>{directionLabel[c.direction] ?? c.direction} · <Link to={docPath(directionSection(c.direction))}>o que significa</Link></dd>
              <dt>Transformação</dt><dd>{c.transformation}</dd>
              <dt>Fórmula / método</dt><dd>{methodLabel[c.method] ?? c.method}: {c.formula} · <Link to={docPath(docSection.normalization)}>normalização</Link></dd>
              <dt>Peso</dt><dd>{fmtPercent(c.weight, 1)} (bruto {formatNumber(c.rawWeight)}) · {c.source.parameter}</dd>
              <dt>Resultado</dt><dd>{fmt2(c.score)} × {fmtPercent(c.weight, 1)} = <strong>{fmt2(c.contribution)}</strong></dd>
            </dl>
          )}
          {c.warnings.map((w, i) => <p key={i} className="small warn-line">{w}</p>)}
        </>
      )}
    </article>
  )
}
