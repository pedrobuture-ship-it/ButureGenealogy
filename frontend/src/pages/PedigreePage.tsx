import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { EmptyState } from '../components/EmptyState'
import { PageHeader } from '../components/PageHeader'
import { PedigreeView } from '../features/pedigree/PedigreeView'
import { useAsync } from '../hooks/useAsync'
import { useDebounced } from '../hooks/useDebounced'
import { api } from '../utils/api'
import { animalTitle } from '../utils/format'

export function PedigreePage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const q = useDebounced(search)

  const animal = useAsync(async () => (id ? api.getAnimal(id) : null))
  const [lastId, setLastId] = useState(id)
  if (lastId !== id) { setLastId(id); animal.reload() }

  const results = useAsync(() => (q.trim() ? api.listAnimals({ search: q, pageSize: 8, includeArchived: true }) : Promise.resolve(null)))
  const [lastQ, setLastQ] = useState(q)
  if (lastQ !== q) { setLastQ(q); results.reload() }

  return (
    <>
      <PageHeader title="Genealogia" subtitle={animal.data ? animalTitle(animal.data) : 'Pedigree e ancestrais de um animal'} />
      <div className="card block">
        <label htmlFor="ped-search" className="field-label">Buscar animal</label>
        <input id="ped-search" type="search" placeholder="Nome ou registro" value={search} onChange={(e) => setSearch(e.target.value)} />
        {results.data && (
          results.data.items.length === 0 ? <p className="muted">Nenhum animal encontrado.</p> : (
            <ul className="plain-list picker" aria-label="Resultados da busca">
              {results.data.items.map((a) => (
                <li key={a.id}>
                  <button className="btn btn-link" onClick={() => { navigate(`/genealogia/${a.id}`); setSearch('') }}>
                    {animalTitle(a)} <span className="muted mono">{a.registration}</span>
                  </button>
                </li>
              ))}
            </ul>
          )
        )}
      </div>
      {id && animal.error && <div role="alert" className="notice notice-error">{animal.error}</div>}
      {id && animal.data ? <PedigreeView animalId={id} /> : !id && (
        <EmptyState title="Escolha um animal">Busque um animal acima ou abra a aba Genealogia na ficha dele para ver o pedigree.</EmptyState>
      )}
    </>
  )
}
