import { Drawer } from '../../components/Drawer'
import { useAsync } from '../../hooks/useAsync'
import type { PairSnapshot } from '../../types/engine'
import { api } from '../../utils/api'
import { CalculationView } from './CalculationView'

interface Props {
  /** A saved calculation to load, or an in-memory snapshot (unsaved preview). */
  matingId?: string
  snapshot?: PairSnapshot
  onClose: () => void
}

export function CalculationDrawer({ matingId, snapshot, onClose }: Props) {
  const saved = useAsync(async () => (matingId && !snapshot ? api.getCalculation(matingId) : null))
  const snap = snapshot ?? saved.data?.snapshot ?? null
  return (
    <Drawer title="Ver cálculo" onClose={onClose}>
      {saved.loading && !snap && <p role="status" className="muted">Carregando cálculo salvo…</p>}
      {saved.error && <div role="alert" className="notice notice-error">{saved.error}</div>}
      {snap && <CalculationView snapshot={snap} />}
    </Drawer>
  )
}
