import type { Provenance } from '../types/domain'
import { formatDate, formatNumber } from '../utils/format'
import { DemoBadge } from './Badge'

interface Props {
  provenance: Provenance
  /** "inline": expandable "Origem" disclosure for table rows. "block": always visible. */
  variant?: 'inline' | 'block'
}

/**
 * Shows where a data point came from. It never merges kinds: a DEP, an EBV and
 * a phenotype are always labelled as what they are.
 */
export function DataProvenance({ provenance: p, variant = 'inline' }: Props) {
  const rows: [string, string][] = [
    ['Característica', p.traitName || p.traitCode],
    ['Tipo', p.kindLabel],
  ]
  if (p.genetic) {
    if (p.datasetName) rows.push(['Dataset', p.datasetName])
    if (p.datasetVersion) rows.push(['Versão', p.datasetVersion])
    if (p.program) rows.push(['Programa', p.program])
    if (p.release) rows.push(['Release da avaliação', p.release])
    if (p.geneticBase) rows.push(['Base genética', p.geneticBase])
    if (p.accuracy != null) rows.push(['Acurácia', `${formatNumber(p.accuracy * 100, 0)}%`])
    if (p.sourceReference) rows.push(['Referência', p.sourceReference])
  } else if (p.source) {
    rows.push(['Origem da medição', p.source])
  }
  if (p.date) rows.push(['Data', formatDate(p.date)])

  const body = (
    <dl className="kv provenance-kv">
      {rows.map(([k, v]) => (<span key={k} className="kv-row"><dt>{k}</dt><dd>{v}</dd></span>))}
      {p.isDemo && <span className="kv-row"><dt>Aviso</dt><dd><DemoBadge /> dado fictício (DEMO / SYNTHETIC)</dd></span>}
    </dl>
  )
  if (variant === 'block') return <div className="provenance">{body}</div>
  return (
    <details className="provenance">
      <summary>Origem</summary>
      {body}
    </details>
  )
}
