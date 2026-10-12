import { EmptyState } from '../../components/EmptyState'
import { useAsync } from '../../hooks/useAsync'
import { api } from '../../utils/api'
import { TraitGroups } from './TraitGroups'

export function GeneticsTab({ animalId }: { animalId: string }) {
  const traits = useAsync(() => api.getAnimalTraits(animalId))

  if (traits.loading) return <p role="status" className="muted">Carregando avaliações…</p>
  if (traits.error) return <div role="alert" className="notice notice-error">{traits.error}</div>
  const groups = traits.data ?? []
  if (!groups.some((g) => g.items.some((i) => i.source === 'evaluation'))) {
    return (
      <EmptyState title="Sem avaliações genéticas">
        Avaliações (DEP, EPD, EBV, PTA, índices) aparecem aqui quando importadas de uma biblioteca na página Bibliotecas.
        Valores de tipos ou programas diferentes nunca são somados nem convertidos automaticamente.
      </EmptyState>
    )
  }
  return <TraitGroups groups={groups} source="evaluation" />
}
