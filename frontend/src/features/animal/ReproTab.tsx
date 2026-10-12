import { EmptyState } from '../../components/EmptyState'
import type { AnimalDetail } from '../../types/domain'

export function ReproTab({ detail }: { detail: AnimalDetail }) {
  return (
    <>
      <div className="grid">
        <div className="card">
          <h2>Resumo reprodutivo</h2>
          <dl className="kv">
            <dt>Descendentes cadastrados</dt><dd>{detail.offspring.length}</dd>
            {detail.animal.sex === 'male' && (<><dt>Doses de sêmen em estoque</dt><dd>{detail.semenDoses}</dd></>)}
          </dl>
        </div>
      </div>
      <EmptyState title="Histórico reprodutivo em desenvolvimento">
        Coberturas, diagnósticos de gestação e partos serão registrados aqui em uma etapa futura.
      </EmptyState>
    </>
  )
}
