import { useState } from 'react'
import { useAsync } from '../../hooks/useAsync'
import { useDebounced } from '../../hooks/useDebounced'
import type { Animal } from '../../types/domain'
import { api } from '../../utils/api'
import { animalTitle } from '../../utils/format'

interface Props {
  selected: Animal | null
  onSelect: (a: Animal | null) => void
}

export function FemalePicker({ selected, onSelect }: Props) {
  const [q, setQ] = useState('')
  const dq = useDebounced(q)
  const list = useAsync(() => api.listAnimals({ sex: 'female', search: dq, pageSize: 8, sortBy: 'name' }))
  const [last, setLast] = useState(dq)
  if (last !== dq) { setLast(dq); list.reload() }

  if (selected) {
    return (
      <div className="selected-chip">
        <div>
          <strong>{animalTitle(selected)}</strong>
          {selected.registration && selected.name && <span className="muted mono"> {selected.registration}</span>}
        </div>
        <button className="btn btn-link" onClick={() => onSelect(null)}>Trocar</button>
      </div>
    )
  }
  return (
    <div>
      <label htmlFor="female-search" className="field-label">Buscar matriz</label>
      <input id="female-search" type="search" placeholder="Nome ou registro" value={q} onChange={(e) => setQ(e.target.value)} />
      {list.data && (list.data.items.length === 0
        ? <p className="muted">Nenhuma fêmea encontrada.</p>
        : <ul className="plain-list picker" aria-label="Fêmeas encontradas">
            {list.data.items.map((a) => (
              <li key={a.id}><button className="btn btn-link" onClick={() => onSelect(a)}>{animalTitle(a)} <span className="muted mono">{a.registration}</span></button></li>
            ))}
          </ul>)}
    </div>
  )
}
