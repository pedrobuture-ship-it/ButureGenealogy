import { useState } from 'react'
import { Modal } from '../../components/Modal'
import { useAsync } from '../../hooks/useAsync'
import { useDebounced } from '../../hooks/useDebounced'
import { api } from '../../utils/api'
import { animalTitle } from '../../utils/format'

interface Props {
  onClose: () => void
  onPick: (bullId: string) => void
  onCreateBull: () => void
}

/** First step of "add semen": find an existing bull, or register a new one. */
export function AddSemenModal({ onClose, onPick, onCreateBull }: Props) {
  const [q, setQ] = useState('')
  const dq = useDebounced(q)
  const results = useAsync(() => (dq.trim() ? api.listAnimals({ search: dq, sex: 'male', includeArchived: false, pageSize: 10 }) : Promise.resolve(null)))
  const [last, setLast] = useState(dq)
  if (last !== dq) { setLast(dq); results.reload() }

  return (
    <Modal title="Adicionar sêmen" onClose={onClose}
      footer={<button className="btn" onClick={onClose}>Fechar</button>}>
      <label htmlFor="bull-search" className="field-label">Buscar touro cadastrado</label>
      <input id="bull-search" type="search" autoFocus placeholder="Nome ou registro" value={q} onChange={(e) => setQ(e.target.value)} />
      {results.data && (results.data.items.length === 0
        ? <p className="muted">Nenhum macho encontrado com esse nome ou registro.</p>
        : <ul className="plain-list picker" aria-label="Touros encontrados">
            {results.data.items.map((a) => (
              <li key={a.id}><button className="btn btn-link" onClick={() => onPick(a.id)}>{animalTitle(a)} <span className="muted mono">{a.registration}</span></button></li>
            ))}
          </ul>)}
      <p className="muted spaced">Não encontrou? <button className="btn btn-link" onClick={onCreateBull}>Cadastrar novo touro</button></p>
    </Modal>
  )
}
