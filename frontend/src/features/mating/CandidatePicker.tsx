import { useEffect, useState } from 'react'
import { useDebounced } from '../../hooks/useDebounced'
import type { BankPage } from '../../types/domain'
import { api, errorMessage } from '../../utils/api'
import { formatMoney } from '../../utils/format'
import { StockBadge } from '../semen/StockBadge'

interface Props {
  selected: Set<string>
  onChange: (next: Set<string>) => void
}

export function CandidatePicker({ selected, onChange }: Props) {
  const [search, setSearch] = useState('')
  const [withoutStock, setWithoutStock] = useState(false)
  const dq = useDebounced(search)
  const [page, setPage] = useState<BankPage | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    api.semenBank({ search: dq, stockFilter: withoutStock ? 'all' : 'in_stock', pageSize: 200, sortBy: 'name' })
      .then((p) => { if (!cancelled) { setPage(p); setError(null) } })
      .catch((e: unknown) => { if (!cancelled) setError(errorMessage(e)) })
    return () => { cancelled = true }
  }, [dq, withoutStock])

  const items = page?.items ?? []
  const toggle = (id: string) => {
    const next = new Set(selected)
    if (next.has(id)) next.delete(id); else next.add(id)
    onChange(next)
  }
  const allSelected = items.length > 0 && items.every((b) => selected.has(b.id))

  return (
    <div>
      <div className="row-gap">
        <input type="search" aria-label="Buscar touro candidato" placeholder="Buscar touro" className="grow" value={search} onChange={(e) => setSearch(e.target.value)} />
        <label className="check"><input type="checkbox" checked={withoutStock} onChange={(e) => setWithoutStock(e.target.checked)} /> Incluir touros sem estoque</label>
      </div>
      {error && <div role="alert" className="notice notice-error">{error}</div>}
      {page && items.length === 0 && <p className="muted">Nenhum touro disponível. Cadastre touros e sêmen na página Sêmen.</p>}
      {items.length > 0 && (
        <div className="table-wrap candidate-list">
          <table className="data-table compact">
            <caption className="sr-only">Touros candidatos</caption>
            <thead>
              <tr>
                <th><input type="checkbox" aria-label="Selecionar todos os touros listados" checked={allSelected}
                  onChange={() => onChange(allSelected ? new Set([...selected].filter((id) => !items.some((b) => b.id === id))) : new Set([...selected, ...items.map((b) => b.id)]))} /></th>
                <th>Touro</th><th>Raça</th><th className="right">Doses</th><th className="right">Custo/dose</th><th>Status</th>
              </tr>
            </thead>
            <tbody>
              {items.map((b) => (
                <tr key={b.id}>
                  <td><input type="checkbox" aria-label={`Selecionar ${b.name || b.registration}`} checked={selected.has(b.id)} onChange={() => toggle(b.id)} /></td>
                  <td><strong>{b.name || b.registration}</strong>{b.isDemo && <span className="badge badge-demo"> DEMO</span>}</td>
                  <td>{b.breedName || '—'}</td><td className="right">{b.doses}</td><td className="right">{formatMoney(b.unitCost)}</td><td><StockBadge status={b.stockStatus} /></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <p className="muted small">{selected.size} touro(s) selecionado(s)</p>
    </div>
  )
}
