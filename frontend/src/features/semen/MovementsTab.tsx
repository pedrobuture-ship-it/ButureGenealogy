import { EmptyState } from '../../components/EmptyState'
import type { BullSheet } from '../../types/domain'
import { formatDateTime } from '../../utils/format'
import { movementTypeLabel, reasonLabel } from '../../utils/labels'

export function MovementsTab({ sheet }: { sheet: BullSheet }) {
  if (sheet.movements.length === 0) {
    return <EmptyState title="Nenhuma movimentação">Entradas, retiradas e ajustes de doses aparecem aqui, com o saldo após cada operação.</EmptyState>
  }
  const lotName = (id: string) => {
    const l = sheet.lots.find((x) => x.id === id)
    return l ? `${l.supplier || 'Sem fornecedor'}${l.batch ? ` · ${l.batch}` : ''}` : '—'
  }
  return (
    <div className="table-wrap">
      <table className="data-table compact">
        <caption className="sr-only">Histórico de movimentações de estoque</caption>
        <thead><tr><th>Data</th><th>Tipo</th><th>Lote</th><th className="right">Doses</th><th className="right">Saldo do lote</th><th>Motivo</th><th>Observação</th></tr></thead>
        <tbody>
          {sheet.movements.map((m) => (
            <tr key={m.id}>
              <td>{formatDateTime(m.createdAt)}</td>
              <td><span className={m.type === 'entry' ? 'badge badge-ok' : m.type === 'withdrawal' ? 'badge badge-warn' : 'badge'}>{movementTypeLabel[m.type]}</span></td>
              <td>{lotName(m.inventoryId)}</td>
              <td className="right"><strong>{m.delta > 0 ? `+${m.delta}` : m.delta}</strong></td>
              <td className="right">{m.balanceAfter}</td>
              <td>{reasonLabel(m.reason)}</td>
              <td>{m.note || '—'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
