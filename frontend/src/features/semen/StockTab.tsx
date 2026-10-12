import { useState } from 'react'
import { EmptyState } from '../../components/EmptyState'
import type { BullSheet, SemenLot } from '../../types/domain'
import { formatMoney } from '../../utils/format'
import { StockBadge } from './StockBadge'
import { LotModal } from './LotModal'
import { MovementModal, type MovementKind } from './MovementModal'

interface Props {
  sheet: BullSheet
  typeName: (code: string) => string
  openNewLot?: boolean
  onChanged: () => void
}

export function StockTab({ sheet, typeName, openNewLot, onChanged }: Props) {
  const [lotModal, setLotModal] = useState<null | { lot?: SemenLot }>(openNewLot ? {} : null)
  const [move, setMove] = useState<null | { kind: MovementKind; lot?: SemenLot }>(null)
  const done = () => { setLotModal(null); setMove(null); onChanged() }

  return (
    <>
      <div className="grid stock-summary">
        <div className="card"><h2>Doses em estoque</h2><p className="big-number" aria-label="Doses em estoque">{sheet.totalDoses.toLocaleString('pt-BR')}</p><StockBadge status={sheet.stockStatus} /></div>
        <div className="card"><h2>Custo médio por dose</h2><p className="big-number">{formatMoney(sheet.unitCost)}</p>
          <p className="muted small">Média ponderada pelas doses dos lotes com preço.</p></div>
        <div className="card"><h2>Lotes</h2><p className="big-number">{sheet.lots.length}</p></div>
      </div>

      <div className="tab-actions">
        <button className="btn btn-primary" onClick={() => setLotModal({})}>Novo lote</button>
      </div>

      {sheet.lots.length === 0 ? (
        <EmptyState title="Nenhum lote registrado">Cadastre um lote para registrar fornecedor, tipo, preço e doses deste touro.</EmptyState>
      ) : (
        <div className="table-wrap">
          <table className="data-table compact">
            <caption className="sr-only">Lotes de sêmen</caption>
            <thead><tr><th>Fornecedor</th><th>Lote</th><th>Tipo</th><th className="right">Doses</th><th className="right">Preço/dose</th><th>Observações</th><th><span className="sr-only">Ações</span></th></tr></thead>
            <tbody>
              {sheet.lots.map((l) => (
                <tr key={l.id}>
                  <td>{l.supplier || '—'}</td><td>{l.batch || '—'}</td><td>{typeName(l.semenType)}</td>
                  <td className="right"><strong>{l.quantity}</strong></td><td className="right">{formatMoney(l.unitCost)}</td><td>{l.notes || '—'}</td>
                  <td className="actions-cell">
                    <button className="btn btn-link" onClick={() => setMove({ kind: 'entry', lot: l })}>Adicionar doses</button>
                    <button className="btn btn-link" onClick={() => setMove({ kind: 'withdrawal', lot: l })} disabled={l.quantity === 0}>Retirar</button>
                    <button className="btn btn-link" onClick={() => setMove({ kind: 'adjustment', lot: l })}>Ajustar</button>
                    <button className="btn btn-link" onClick={() => setLotModal({ lot: l })}>Editar</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {lotModal && <LotModal bullId={sheet.detail.animal.id} lot={lotModal.lot} onClose={() => setLotModal(null)} onSaved={done} />}
      {move && <MovementModal kind={move.kind} lots={sheet.lots} lot={move.lot} onClose={() => setMove(null)} onSaved={done} />}
    </>
  )
}
