import { DemoBadge } from '../../components/Badge'
import type { CalcStatus, CandidateResult, RunOutput } from '../../types/engine'
import { animalTitle, fmt2, fmtPercent } from '../../utils/format'

interface Props {
  output: RunOutput
  saved: boolean
  onViewCalculation: (bullId: string) => void
}

const label = (a: { name: string; registration: string }) => animalTitle(a)

export function ResultsView({ output, saved, onViewCalculation }: Props) {
  const calc = output.results.filter((r) => r.status === 'calculated')
  const notCalc = output.results.filter((r) => r.status === 'not_calculable')
  const excluded = output.results.filter((r) => r.status === 'excluded')
  const top = output.top.map((id) => calc.find((r) => r.bull.id === id)).filter((r): r is CandidateResult => !!r).slice(0, 3)

  if (output.errors.length > 0) {
    return (
      <div role="alert" className="notice notice-error">
        <strong>Não foi possível calcular</strong>
        <ul className="plain-list">{output.errors.map((e, i) => <li key={i}>{e}</li>)}</ul>
      </div>
    )
  }

  return (
    <div className="results">
      {output.profile.isDemo && <div className="notice notice-warn" role="note"><strong>{output.profile.demoNotice}</strong></div>}
      <p className="muted small">Motor {output.engineVersion} · perfil {output.profile.name}{saved ? ' · resultado salvo no histórico' : ' · não salvo'}</p>

      {top.length > 0 && (
        <section aria-label="Top candidatos" className="top-cards">
          {top.map((r) => (
            <article key={r.bull.id} className="card top-card">
              <p className="rank-badge">#{r.rank}</p>
              <h3>{label(r.bull)} {r.bull.isDemo && <DemoBadge />}</h3>
              <p className="top-score">{fmt2(r.final?.final)} <span className="muted">/ 100</span></p>
              <p className="muted small">{r.bull.breedName}{r.final ? ` · cobertura ${fmtPercent(r.final.coverage)}` : ''}</p>
              <button className="btn" onClick={() => onViewCalculation(r.bull.id)}>Ver cálculo</button>
            </article>
          ))}
        </section>
      )}

      <section aria-label="Ranking">
        <h2>Ranking</h2>
        {calc.length === 0 ? <p className="muted">Nenhum touro pôde ter o score calculado com os dados e o perfil atuais.</p> : (
          <div className="table-wrap">
            <table className="data-table">
              <caption className="sr-only">Ranking de touros</caption>
              <thead>
                <tr>
                  <th>#</th><th>Touro</th>
                  {calc[0].components.map((c) => <th key={c.traitCode + c.evaluationType} className="right" title={c.name}>{c.traitCode}</th>)}
                  <th className="right">Compl.</th><th className="right">Penal.</th><th className="right">Score</th><th>Avisos</th><th><span className="sr-only">Ações</span></th>
                </tr>
              </thead>
              <tbody>
                {calc.map((r) => (
                  <tr key={r.bull.id}>
                    <td>{r.rank}{r.tied && <span title="Empate"> =</span>}</td>
                    <td><strong>{label(r.bull)}</strong> {r.bull.isDemo && <DemoBadge />}</td>
                    {r.components.map((c) => <td key={c.traitCode + c.evaluationType} className="right">{c.status === 'calculated' ? fmt2(c.score) : <span className="muted" title="Não calculável">—</span>}</td>)}
                    <td className="right">{r.final && r.final.complementarity > 0 ? `+${fmt2(r.final.complementarity)}` : '—'}</td>
                    <td className="right">{r.final && r.final.penalties > 0 ? `−${fmt2(r.final.penalties)}` : '—'}</td>
                    <td className="right"><strong>{fmt2(r.final?.final)}</strong></td>
                    <td>{r.warnings.length > 0 ? <span className="badge badge-warn" title={r.warnings.join('\n')}>{r.warnings.length}</span> : '—'}</td>
                    <td><button className="btn btn-link" onClick={() => onViewCalculation(r.bull.id)}>Ver cálculo</button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {notCalc.length > 0 && (
        <section aria-label="Não calculáveis">
          <h2>Não calculáveis</h2>
          <p className="muted small">Estes touros não receberam score: preferimos “não calculável” a um número sem base.</p>
          {notCalc.map((r) => <Reasons key={r.bull.id} r={r} onView={onViewCalculation} />)}
        </section>
      )}

      {excluded.length > 0 && (
        <section aria-label="Excluídos pelos filtros">
          <h2>Excluídos pelos filtros</h2>
          {excluded.map((r) => (
            <div key={r.bull.id} className="reason-row">
              <strong>{label(r.bull)}</strong> <span className="badge badge-muted">excluído</span>
              <ul className="plain-list small">{r.exclusions.map((x, i) => <li key={i}><em>{x.filter}:</em> {x.reason}</li>)}</ul>
            </div>
          ))}
        </section>
      )}

      {output.bases.length > 0 && (
        <details className="card block">
          <summary>Bases de comparação usadas</summary>
          <ul className="plain-list small">
            {output.bases.map((b) => (
              <li key={b.traitCode + b.evaluationType}>
                <strong>{b.traitCode}</strong>: {b.basis ? `${b.basis.program} · ${b.basis.release} · ${b.basis.geneticBase || 'base não informada'}` : 'nenhuma base disponível'} — {b.reason}
              </li>
            ))}
          </ul>
        </details>
      )}
      {output.warnings.length > 0 && (
        <div className="notice notice-warn"><strong>Avisos da execução</strong><ul className="plain-list">{output.warnings.map((w, i) => <li key={i}>{w}</li>)}</ul></div>
      )}
    </div>
  )
}

const statusText: Record<CalcStatus, string> = { calculated: 'calculado', not_calculable: 'não calculável', not_applied: 'não aplicado', excluded: 'excluído' }

function Reasons({ r, onView }: { r: CandidateResult; onView: (id: string) => void }) {
  const incompatible = r.components.filter((c) => c.warnings.some((w) => w.startsWith('⚠')))
  const others = r.components.filter((c) => c.status !== 'calculated' && !incompatible.includes(c)).map((c) => `${c.name}: ${c.explanation}`)
  const items = [...r.reasons, ...others]
  return (
    <div className="reason-row">
      <strong>{label(r.bull)}</strong> <span className="badge badge-warn">{statusText[r.status]}</span>{' '}
      <button className="btn btn-link" onClick={() => onView(r.bull.id)}>Ver cálculo</button>
      <ul className="plain-list small">
        {items.filter((x, i, a) => a.indexOf(x) === i).map((x, i) => <li key={i}>{x}</li>)}
        {incompatible.length > 0 && (
          <li className="warn-line">⚠ Comparação genética não suportada entre essas bases ({incompatible.map((c) => c.traitCode).join(', ')}): as avaliações deste touro vêm de outro programa, release ou base e não são convertidas.</li>
        )}
      </ul>
    </div>
  )
}
