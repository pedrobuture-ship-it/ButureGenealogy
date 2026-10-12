import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { DemoBadge, StatusBadge } from '../components/Badge'
import { EmptyState } from '../components/EmptyState'
import { PageHeader } from '../components/PageHeader'
import { useToast } from '../components/Toast'
import { AnimalFormModal } from '../features/herd/AnimalFormModal'
import { FarmFormModal } from '../features/herd/FarmFormModal'
import { useAsync } from '../hooks/useAsync'
import { useDebounced } from '../hooks/useDebounced'
import type { AnimalPage, AnimalQuery } from '../types/domain'
import { api, errorMessage } from '../utils/api'
import { animalTitle, formatAge, formatDate } from '../utils/format'
import { categoryLabel, sexLabel, statusLabel } from '../utils/labels'

const PAGE_SIZE = 50

const ageRanges: { id: string; label: string; min: number | null; max: number | null }[] = [
  { id: '', label: 'Qualquer idade', min: null, max: null },
  { id: 'lt12', label: 'Até 12 meses', min: null, max: 12 },
  { id: '12-24', label: '12 a 24 meses', min: 12, max: 24 },
  { id: '24-48', label: '2 a 4 anos', min: 24, max: 48 },
  { id: 'gt48', label: 'Mais de 4 anos', min: 48, max: null },
]

interface Filters {
  farmId: string
  breedId: string
  sex: string
  category: string
  status: string // '' = ativos, 'all' = todos incl. arquivados
  age: string
  search: string
}
const emptyFilters: Filters = { farmId: '', breedId: '', sex: '', category: '', status: '', age: '', search: '' }

const columns: { key: string; label: string }[] = [
  { key: 'registration', label: 'Identificação' },
  { key: 'name', label: 'Nome' },
  { key: 'breed', label: 'Raça' },
  { key: 'sex', label: 'Sexo' },
  { key: 'birthDate', label: 'Nascimento / idade' },
  { key: 'category', label: 'Categoria' },
  { key: 'status', label: 'Status' },
]

export function HerdPage() {
  const navigate = useNavigate()
  const toast = useToast()
  const refs = useAsync(async () => {
    const [farms, breeds] = await Promise.all([api.listFarms(), api.listBreeds()])
    return { farms, breeds }
  })

  const [filters, setFilters] = useState<Filters>(emptyFilters)
  const [sort, setSort] = useState({ by: 'name', desc: false })
  const [page, setPage] = useState(1)
  const [result, setResult] = useState<AnimalPage | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [reloadTick, setReloadTick] = useState(0)
  const [modal, setModal] = useState<null | 'farm' | 'animal'>(null)
  const [busyDemo, setBusyDemo] = useState(false)

  const debouncedSearch = useDebounced(filters.search)

  const query = useMemo<Partial<AnimalQuery>>(() => {
    const age = ageRanges.find((a) => a.id === filters.age)
    return {
      farmId: filters.farmId, breedId: filters.breedId, sex: filters.sex, category: filters.category,
      status: filters.status === 'all' ? '' : filters.status,
      includeArchived: filters.status === 'all',
      search: debouncedSearch, minAgeMonths: age?.min ?? null, maxAgeMonths: age?.max ?? null,
      sortBy: sort.by, sortDesc: sort.desc, page, pageSize: PAGE_SIZE,
    }
  }, [filters, debouncedSearch, sort, page])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    api.listAnimals(query)
      .then((r) => { if (!cancelled) { setResult(r); setError(null) } })
      .catch((e: unknown) => { if (!cancelled) setError(errorMessage(e)) })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [query, reloadTick])

  const reload = useCallback(() => setReloadTick((t) => t + 1), [])
  const setFilter = (k: keyof Filters, v: string) => { setFilters((f) => ({ ...f, [k]: v })); setPage(1) }
  const filtersActive = JSON.stringify(filters) !== JSON.stringify(emptyFilters)

  function toggleSort(key: string) {
    setSort((s) => (s.by === key ? { by: key, desc: !s.desc } : { by: key, desc: false }))
    setPage(1)
  }

  async function loadDemo() {
    setBusyDemo(true)
    try {
      await api.loadDemoData()
      toast('success', 'Dados DEMO / SYNTHETIC carregados.')
      refs.reload()
      reload()
    } catch (e) {
      toast('error', errorMessage(e))
    } finally {
      setBusyDemo(false)
    }
  }

  const farms = refs.data?.farms ?? []
  const breeds = refs.data?.breeds ?? []
  const breedName = (id: string) => breeds.find((b) => b.id === id)?.name ?? '—'
  const totalPages = result ? Math.max(1, Math.ceil(result.total / PAGE_SIZE)) : 1

  if (refs.loading) return <p role="status" className="muted">Carregando…</p>
  if (refs.error) return <div role="alert" className="notice notice-error">{refs.error}</div>

  const actions = (
    <>
      <button className="btn" onClick={() => setModal('farm')}>Nova propriedade</button>
      <button className="btn btn-primary" onClick={() => setModal('animal')} disabled={farms.length === 0}
        title={farms.length === 0 ? 'Cadastre uma propriedade primeiro' : undefined}>
        Novo animal
      </button>
    </>
  )

  return (
    <>
      <PageHeader title="Rebanho" subtitle="Cadastro e consulta dos animais" actions={actions} />

      {farms.length === 0 ? (
        <EmptyState title="Nenhuma propriedade cadastrada">
          Cadastre sua propriedade para começar a registrar animais. Para apenas explorar o aplicativo, você pode carregar dados fictícios (DEMO / SYNTHETIC).
          <span className="empty-actions">
            <button className="btn btn-primary" onClick={() => setModal('farm')}>Cadastrar propriedade</button>
            <button className="btn" onClick={loadDemo} disabled={busyDemo}>{busyDemo ? 'Carregando…' : 'Carregar dados de demonstração'}</button>
          </span>
        </EmptyState>
      ) : (
        <>
          <section className="filters" aria-label="Filtros">
            <input type="search" placeholder="Buscar por nome ou registro" aria-label="Buscar por nome ou registro"
              value={filters.search} onChange={(e) => setFilter('search', e.target.value)} className="filter-search" />
            <select aria-label="Propriedade" value={filters.farmId} onChange={(e) => setFilter('farmId', e.target.value)}>
              <option value="">Todas as propriedades</option>
              {farms.map((f) => <option key={f.id} value={f.id}>{f.name}</option>)}
            </select>
            <select aria-label="Raça" value={filters.breedId} onChange={(e) => setFilter('breedId', e.target.value)}>
              <option value="">Todas as raças</option>
              {breeds.map((b) => <option key={b.id} value={b.id}>{b.name}</option>)}
            </select>
            <select aria-label="Sexo" value={filters.sex} onChange={(e) => setFilter('sex', e.target.value)}>
              <option value="">Ambos os sexos</option>
              {Object.entries(sexLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
            </select>
            <select aria-label="Categoria" value={filters.category} onChange={(e) => setFilter('category', e.target.value)}>
              <option value="">Todas as categorias</option>
              {Object.entries(categoryLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
            </select>
            <select aria-label="Status" value={filters.status} onChange={(e) => setFilter('status', e.target.value)}>
              <option value="">Somente ativos</option>
              <option value="all">Todos (inclui arquivados)</option>
              {Object.entries(statusLabel).map(([v, l]) => <option key={v} value={v}>{l}</option>)}
            </select>
            <select aria-label="Idade aproximada" value={filters.age} onChange={(e) => setFilter('age', e.target.value)}>
              {ageRanges.map((a) => <option key={a.id} value={a.id}>{a.label}</option>)}
            </select>
            {filtersActive && <button className="btn btn-link" onClick={() => { setFilters(emptyFilters); setPage(1) }}>Limpar filtros</button>}
          </section>

          {error && <div role="alert" className="notice notice-error">{error}</div>}

          <div className="table-wrap" aria-busy={loading}>
            <table className="data-table">
              <caption className="sr-only">Animais do rebanho</caption>
              <thead>
                <tr>
                  {columns.map((c) => (
                    <th key={c.key} scope="col" aria-sort={sort.by === c.key ? (sort.desc ? 'descending' : 'ascending') : 'none'}>
                      <button className="th-btn" onClick={() => toggleSort(c.key)}>
                        {c.label}
                        <span aria-hidden="true">{sort.by === c.key ? (sort.desc ? ' ↓' : ' ↑') : ''}</span>
                      </button>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {result?.items.map((a) => (
                  <tr key={a.id} className="row-link" tabIndex={0}
                    onClick={() => navigate(`/rebanho/${a.id}`)}
                    onKeyDown={(e) => { if (e.key === 'Enter') navigate(`/rebanho/${a.id}`) }}>
                    <td className="mono">{a.registration || '—'}</td>
                    <td><strong>{animalTitle(a)}</strong> {a.isDemo && <DemoBadge />}</td>
                    <td>{breedName(a.breedId)}</td>
                    <td>{sexLabel[a.sex]}</td>
                    <td>{a.birthDate ? `${formatDate(a.birthDate)} · ${formatAge(a.birthDate)}` : '—'}</td>
                    <td>{categoryLabel[a.category]}</td>
                    <td><StatusBadge status={a.status} label={statusLabel[a.status]} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
            {loading && !result && <p role="status" className="muted pad">Carregando animais…</p>}
            {result && result.items.length === 0 && !loading && (
              <EmptyState title={filtersActive ? 'Nenhum animal encontrado' : 'Nenhum animal cadastrado'}>
                {filtersActive ? 'Nenhum animal corresponde aos filtros atuais.' : 'Use “Novo animal” para cadastrar o primeiro.'}
              </EmptyState>
            )}
          </div>

          {result && result.total > 0 && (
            <nav className="pager" aria-label="Paginação">
              <span className="muted">{result.total} {result.total === 1 ? 'animal' : 'animais'}</span>
              <span className="pager-controls">
                <button className="btn" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>Anterior</button>
                <span>Página {page} de {totalPages}</span>
                <button className="btn" disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)}>Próxima</button>
              </span>
            </nav>
          )}
        </>
      )}

      {modal === 'farm' && (
        <FarmFormModal onClose={() => setModal(null)} onSaved={() => { setModal(null); refs.reload() }} />
      )}
      {modal === 'animal' && (
        <AnimalFormModal farms={farms} breeds={breeds} defaultFarmId={filters.farmId || undefined}
          onClose={() => setModal(null)}
          onSaved={(a) => { setModal(null); navigate(`/rebanho/${a.id}`) }} />
      )}
    </>
  )
}
