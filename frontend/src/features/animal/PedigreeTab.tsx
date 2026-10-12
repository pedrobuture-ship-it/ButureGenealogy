import { Link } from 'react-router-dom'
import type { AnimalDetail, AnimalRef } from '../../types/domain'
import { PedigreeView } from '../pedigree/PedigreeView'
import { sexLabel } from '../../utils/labels'

function RefRow({ role, a }: { role: string; a: AnimalRef | null }) {
  return (
    <>
      <dt>{role}</dt>
      <dd>
        {a ? <Link to={`/rebanho/${a.id}`}>{a.name || a.registration}</Link> : <span className="muted">Não informado</span>}
        {a && a.registration && a.name ? <span className="muted mono"> {a.registration}</span> : null}
      </dd>
    </>
  )
}

export function PedigreeTab({ detail }: { detail: AnimalDetail }) {
  return (
    <>
    <PedigreeView animalId={detail.animal.id} />
    <div className="grid">
      <div className="card">
        <h2>Ascendentes diretos</h2>
        <dl className="kv">
          <RefRow role="Pai" a={detail.father} />
          <RefRow role="Mãe" a={detail.mother} />
        </dl>
      </div>
      <div className="card">
        <h2>Descendentes cadastrados</h2>
        {detail.offspring.length === 0 ? (
          <p className="muted">Nenhum descendente cadastrado.</p>
        ) : (
          <ul className="plain-list">
            {detail.offspring.map((o) => (
              <li key={o.id}><Link to={`/rebanho/${o.id}`}>{o.name || o.registration}</Link> <span className="muted">{sexLabel[o.sex]}</span></li>
            ))}
          </ul>
        )}
      </div>
    </div>
    </>
  )
}
