import type { StockStatus } from '../../types/domain'
import { stockStatusLabel } from '../../utils/labels'

const cls: Record<StockStatus, string> = { in_stock: 'badge-ok', depleted: 'badge-warn', no_semen: 'badge-muted' }

export function StockBadge({ status }: { status: StockStatus }) {
  return <span className={`badge ${cls[status]}`}>{stockStatusLabel[status]}</span>
}
