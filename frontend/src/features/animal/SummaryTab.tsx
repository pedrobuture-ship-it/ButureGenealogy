import type { AnimalDetail } from '../../types/domain'
import { formatAge, formatDate } from '../../utils/format'
import { categoryLabel, sexLabel, statusLabel } from '../../utils/labels'

const dash = (v: string) => v || '—'

export function SummaryTab({ detail }: { detail: AnimalDetail }) {
  const a = detail.animal
  return (
    <div className="grid">
      <div className="card">
        <h2>Identificação</h2>
        <dl className="kv">
          <dt>Nome</dt><dd>{dash(a.name)}</dd>
          <dt>Registro</dt><dd className="mono">{dash(a.registration)}</dd>
          <dt>Propriedade</dt><dd>{dash(detail.farmName)}</dd>
          <dt>Raça</dt><dd>{dash(detail.breedName)}</dd>
          <dt>Sexo</dt><dd>{sexLabel[a.sex]}</dd>
          <dt>Categoria</dt><dd>{categoryLabel[a.category]}</dd>
          <dt>Status</dt><dd>{statusLabel[a.status]}</dd>
        </dl>
      </div>
      <div className="card">
        <h2>Nascimento</h2>
        <dl className="kv">
          <dt>Data</dt><dd>{formatDate(a.birthDate)}</dd>
          <dt>Idade aproximada</dt><dd>{formatAge(a.birthDate)}</dd>
          <dt>Pai</dt><dd>{detail.father ? detail.father.name || detail.father.registration : '—'}</dd>
          <dt>Mãe</dt><dd>{detail.mother ? detail.mother.name || detail.mother.registration : '—'}</dd>
        </dl>
      </div>
      <div className="card span-all">
        <h2>Observações</h2>
        <p className="prewrap">{a.notes || <span className="muted">Nenhuma observação.</span>}</p>
      </div>
    </div>
  )
}
