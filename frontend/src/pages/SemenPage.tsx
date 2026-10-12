import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { DemoBadge } from '../components/Badge'
import { EmptyState } from '../components/EmptyState'
import { PageHeader } from '../components/PageHeader'
import { AddSemenModal } from '../features/semen/AddSemenModal'
import { StockBadge } from '../features/semen/StockBadge'
import { AnimalFormModal } from '../features/herd/AnimalFormModal'
import { useAsync } from '../hooks/useAsync'
import { useDebounced } from '../hooks/useDebounced'
import type { BankPage, BankQuery } from '../types/domain'
import { api, errorMessage } from '../utils/api'
import { formatMoney } from '../utils/format'
import { statusLabel } from '../utils/labels'

const PAGE_SIZE = 50

const columns: { key: string; label: string; numeric?: boolean }[] = [
  { key: 'name', label: 'Touro' },
  { key: 'breed', label: 'Raça' },
  { key: 'registration', label: 'Registro' },
  { key: 'supplier', label: 'Fornecedor' },
  { key: 'doses', label: 'Doses', numeric: true },
  { key: 'unitCost', label: 'Custo/dose', numeric: true },
  { key: 'status', label: 'Status' },
]

export function SemenPage() {
  const navigate = useNavigate()
  const refs = useAsync(async () => {
    const [farms, breeds] = await Promise.all([api.listFarms(), api.listBreeds()])
    return { farms, breeds }
  })
  const [search, setSearch] = useState('')
  const [breedId, setBreedId] = useState('')
  const [supplier, setSupplier] = useState('')
  const [stockFilter, setStockFilter] = useState<BankQuery['stockFilter']>('with_lots')
  const [sort, setSort] = useState({ by: 'name', desc: false })
  const [page, setPage] = useState(1)
  const [result, setResult] = useState<BankPage | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [modal, setModal] = useState<null | 'add' | 'newBull'>(null)
  const dSearch = useDebounced(search)

  const query = useMemo<Partial<BankQuery>>(() => ({
    search: dSearch, breedId, supplier, stockFilter, sortBy: sort.by, sortDesc: sort.desc, page, pageSize: PAGE_SIZE,
  }), [dSearch, breedId, supplier, stockFilter, sort, page])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    api.semenBank(query)
      .then((r) => { if (!cancelled) { setResult(r); setError(null) } })
      .catch((e: unknown) => { if (!cancelled) setError(errorMessage(e)) })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [query])

  const farms = refs.data?.farms ?? []
  const breeds = refs.data?.breeds ?? []
  const totalPages = result ? Math.max(1, Math.ceil(result.total / PAGE_SIZE)) : 1
  const toggleSort = (key: string) => { setSort((s) => (s.by === key ? { by: key, desc: !s.desc } : { by: key, desc: false })); setPage(1) }
  const reset = (fn: () => void) => { fn(); setPage(1) }

  return (
    <>
      <PageHeader title="Sêmen" subtitle="Touros e estoque de doses"
        actions={<button className="btn btn-primary" onClick={() => setModal('add')} disabled={farms.length === 0}
          title={farms.length === 0 ? 'Cadastre uma propriedade na página Rebanho primeiro' : undefined}>Adicionar sêmen</button>} />

      <section className="filters" aria-label="Filtros">
        <input type="search" aria-label="Buscar touro" placeholder="Buscar por nome ou registro" className="filter-search" value={search} onChange={(e) => reset(() => setSearch(e.target.value))} />
        <select aria-label="Raça" value={breedId} onChange={(e) => reset(() => setBreedId(e.target.value))}>
          <option value="">Todas as raças</option>{breeds.map((b) => <option key={b.id} value={b.id}>{b.name}</option>)}
        </select>
        <select aria-label="Fornecedor" value={supplier} onChange={(e) => reset(() => setSupplier(e.target.value))}>
          <option value="">Todos os fornecedores</option>{(result?.suppliers ?? []).map((s) => <option key={s} value={s}>{s}</option>)}
        </select>
        <select aria-label="Estoque" value={stockFilter} onChange={(e) => reset(() => setStockFilter(e.target.value as BankQuery['stockFilter']))}>
          <option value="with_lots">Touros com sêmen registrado</option>
          <option value="in_stock">Somente com doses em estoque</option>
          <option value="all">Todos os touros</option>
        </select>
      </section>

      {error && <div role="alert" className="notice notice-error">{error}</div>}

      <div className="table-wrap" aria-busy={loading}>
        <table className="data-table">
          <caption className="sr-only">Banco de touros e estoque de sêmen</caption>
          <thead>
            <tr>
              {columns.map((c) => (
                <th key={c.key} scope="col" className={c.numeric ? 'right' : ''} aria-sort={sort.by === c.key ? (sort.desc ? 'descending' : 'ascending') : 'none'}>
                  <button className="th-btn" onClick={() => toggleSort(c.key)}>{c.label}<span aria-hidden="true">{sort.by === c.key ? (sort.desc ? ' ↓' : ' ↑') : ''}</span></button>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {result?.items.map((b) => (
              <tr key={b.id} className="row-link" tabIndex={0} onClick={() => navigate(`/semen/${b.id}`)} onKeyDown={(e) => { if (e.key === 'Enter') navigate(`/semen/${b.id}`) }}>
                <td><strong>{b.name || b.registration}</strong> {b.isDemo && <DemoBadge />}{b.animalStatus !== 'active' && <span className="badge badge-muted"> {statusLabel[b.animalStatus]}</span>}</td>
                <td>{b.breedName || '—'}</td>
                <td className="mono">{b.registration || '—'}</td>
                <td>{b.suppliers.join(', ') || '—'}</td>
                <td className="right"><strong>{b.doses.toLocaleString('pt-BR')}</strong></td>
                <td className="right">{formatMoney(b.unitCost)}</td>
                <td><StockBadge status={b.stockStatus} /></td>
              </tr>
            ))}
          </tbody>
        </table>
        {loading && !result && <p role="status" className="muted pad">Carregando…</p>}
        {result && result.items.length === 0 && !loading && (
          <EmptyState title="Nenhum touro encontrado">
            {stockFilter === 'with_lots' && !search && !breedId && !supplier
              ? 'Ainda não há sêmen registrado. Use “Adicionar sêmen” para escolher um touro e cadastrar um lote.'
              : 'Nenhum touro corresponde aos filtros atuais.'}
          </EmptyState>
        )}
      </div>

      {result && result.total > 0 && (
        <nav className="pager" aria-label="Paginação">
          <span className="muted">{result.total} {result.total === 1 ? 'touro' : 'touros'} · {result.totalDoses.toLocaleString('pt-BR')} doses</span>
          <span className="pager-controls">
            <button className="btn" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>Anterior</button>
            <span>Página {page} de {totalPages}</span>
            <button className="btn" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>Próxima</button>
          </span>
        </nav>
      )}

      {modal === 'add' && <AddSemenModal onClose={() => setModal(null)} onPick={(id) => navigate(`/semen/${id}?novoLote=1`)} onCreateBull={() => setModal('newBull')} />}
      {modal === 'newBull' && (
        <AnimalFormModal farms={farms} breeds={breeds} defaults={{ sex: 'male', category: 'bull' }} onClose={() => setModal(null)}
          onSaved={(a) => { setModal(null); navigate(`/semen/${a.id}?novoLote=1`) }} />
      )}
    </>
  )
}
