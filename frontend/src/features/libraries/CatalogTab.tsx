import { useState } from 'react'
import { EmptyState } from '../../components/EmptyState'
import { useAsync } from '../../hooks/useAsync'
import { useDebounced } from '../../hooks/useDebounced'
import { api } from '../../utils/api'
import { directionLabel, traitCategoryLabel, valueKindLabel } from '../../utils/labels'

export function CatalogTab() {
  const [search, setSearch] = useState('')
  const [kind, setKind] = useState('')
  const q = useDebounced(search)
  const traits = useAsync(() => api.listTraits({ search: q, kind }))
  const [last, setLast] = useState({ q, kind })
  if (last.q !== q || last.kind !== kind) { setLast({ q, kind }); traits.reload() }
  const kinds = useAsync(() => api.listValueKinds())

  return (
    <>
      <section className="filters" aria-label="Filtros do catálogo">
        <input type="search" aria-label="Buscar característica" placeholder="Buscar por código ou nome" value={search} onChange={(e) => setSearch(e.target.value)} className="filter-search" />
        <select aria-label="Tipo de valor" value={kind} onChange={(e) => setKind(e.target.value)}>
          <option value="">Todos os tipos</option>
          {Object.entries(valueKindLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
        </select>
      </section>
      {kinds.data && (
        <details className="card block"><summary>O que significa cada tipo de valor?</summary>
          <dl className="kv">{kinds.data.map((k) => (<span key={k.kind} className="kv-row"><dt>{k.label}</dt><dd>{k.description}</dd></span>))}</dl>
          <p className="muted small">Tipos diferentes não são equivalentes e nunca são somados ou convertidos automaticamente.</p>
        </details>
      )}
      {traits.error && <div role="alert" className="notice notice-error">{traits.error}</div>}
      {traits.data && traits.data.length === 0 && <EmptyState title="Nenhuma característica encontrada" />}
      {traits.data && traits.data.length > 0 && (
        <div className="table-wrap">
          <table className="data-table compact">
            <caption className="sr-only">Catálogo de características</caption>
            <thead><tr><th>Código</th><th>Nome</th><th>Categoria</th><th>Tipo</th><th>Unidade</th><th>Direção</th><th>Origem</th></tr></thead>
            <tbody>
              {traits.data.map((t) => (
                <tr key={t.id}>
                  <td className="mono">{t.code}</td><td>{t.name}</td><td>{traitCategoryLabel[t.category]}</td>
                  <td><span className={t.evaluationType === 'PHENOTYPE' ? 'badge' : 'badge badge-info'}>{valueKindLabel[t.evaluationType]}</span></td>
                  <td>{t.unit || '—'}</td><td>{directionLabel[t.optimizationDirection]}</td>
                  <td>{t.isBuiltin ? 'Nativa' : t.sourceDataset ? 'Biblioteca' : 'Criada por você'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  )
}
