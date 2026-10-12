import type { ReactNode } from 'react'
import { DataProvenance } from '../../components/DataProvenance'
import { DemoBadge } from '../../components/Badge'
import type { TraitGroup, TraitItem } from '../../types/domain'
import { formatNumber } from '../../utils/format'
import { traitCategoryLabel, valueKindLabel } from '../../utils/labels'

interface Props {
  groups: TraitGroup[]
  source: 'measurement' | 'evaluation'
  /** Extra per-row actions (e.g. delete). */
  rowActions?: (item: TraitItem) => ReactNode
}

/** Trait values grouped by category: unit, value, provenance, accuracy, percentile. */
export function TraitGroups({ groups, source, rowActions }: Props) {
  const genetic = source === 'evaluation'
  return (
    <>
      {groups
        .map((g) => ({ ...g, items: g.items.filter((i) => i.source === source) }))
        .filter((g) => g.items.length > 0)
        .map((g) => (
          <section key={g.category} className="card block" aria-label={traitCategoryLabel[g.category]}>
            <h2>{traitCategoryLabel[g.category] ?? g.category}</h2>
            <table className="data-table compact">
              <thead>
                <tr>
                  <th>Característica</th>
                  <th>Tipo</th>
                  <th>Valor</th>
                  <th>Unidade</th>
                  {genetic && <th>Acurácia</th>}
                  {genetic && <th>Percentil</th>}
                  {genetic && <th>Dataset · versão</th>}
                  <th>Origem</th>
                  {rowActions && <th />}
                </tr>
              </thead>
              <tbody>
                {g.items.map((i) => (
                  <tr key={i.id}>
                    <td>
                      <strong>{i.traitName}</strong>{' '}
                      {!i.defined && <span className="badge badge-muted" title="Sem definição no catálogo">sem definição</span>}
                      {i.provenance.isDemo && <DemoBadge />}
                    </td>
                    <td><span className={i.provenance.genetic ? 'badge badge-info' : 'badge'}>{valueKindLabel[i.valueKind]}</span></td>
                    <td><strong>{formatNumber(i.value)}</strong></td>
                    <td>{i.unit || '—'}</td>
                    {genetic && <td>{i.accuracy != null ? `${formatNumber(i.accuracy * 100, 0)}%` : '—'}</td>}
                    {genetic && <td>{i.percentile != null ? formatNumber(i.percentile, 0) : '—'}</td>}
                    {genetic && <td>{sourceText(i)}</td>}
                    <td><DataProvenance provenance={i.provenance} /></td>
                    {rowActions && <td className="right">{rowActions(i)}</td>}
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        ))}
    </>
  )
}

/** Visible "where from": dataset and version when there is one, otherwise program and release. */
function sourceText(i: TraitItem): string {
  const p = i.provenance
  const name = p.datasetName || p.program || '—'
  const ver = p.datasetVersion || p.release
  return ver ? `${name} · ${ver}` : name
}
