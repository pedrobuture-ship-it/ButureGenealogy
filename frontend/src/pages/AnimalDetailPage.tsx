import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { DemoBadge, StatusBadge } from '../components/Badge'
import { ConfirmDialog } from '../components/ConfirmDialog'
import { PageHeader } from '../components/PageHeader'
import { Tabs } from '../components/Tabs'
import { useToast } from '../components/Toast'
import { GeneticsTab } from '../features/animal/GeneticsTab'
import { PedigreeTab } from '../features/animal/PedigreeTab'
import { PhenotypeTab } from '../features/animal/PhenotypeTab'
import { ReproTab } from '../features/animal/ReproTab'
import { SummaryTab } from '../features/animal/SummaryTab'
import { AnimalFormModal } from '../features/herd/AnimalFormModal'
import { useAsync } from '../hooks/useAsync'
import { api, errorMessage } from '../utils/api'
import { animalTitle } from '../utils/format'
import { statusLabel } from '../utils/labels'

const tabs = [
  { id: 'summary', label: 'Resumo' },
  { id: 'phenotype', label: 'Fenótipo' },
  { id: 'genetics', label: 'Genética' },
  { id: 'pedigree', label: 'Genealogia' },
  { id: 'repro', label: 'Reprodução' },
]

export function AnimalDetailPage() {
  const { id = '' } = useParams()
  const toast = useToast()
  const [tab, setTab] = useState('summary')
  const [editing, setEditing] = useState(false)
  const [archiving, setArchiving] = useState(false)

  const detail = useAsync(async () => {
    const [d, farms, breeds] = await Promise.all([api.getAnimalDetail(id), api.listFarms(), api.listBreeds()])
    return { d, farms, breeds }
  })
  // refetch when navigating between animals (e.g. clicking a parent)
  const [lastId, setLastId] = useState(id)
  if (lastId !== id) { setLastId(id); detail.reload() }

  if (detail.loading && !detail.data) return <p role="status" className="muted">Carregando animal…</p>
  if (detail.error || !detail.data) {
    return (
      <>
        <div role="alert" className="notice notice-error">{detail.error ?? 'Animal não encontrado.'}</div>
        <Link to="/rebanho">Voltar ao rebanho</Link>
      </>
    )
  }

  const { d, farms, breeds } = detail.data
  const a = d.animal
  const archived = a.status === 'archived'

  return (
    <>
      <p className="breadcrumb"><Link to="/rebanho">Rebanho</Link> / {animalTitle(a)}</p>
      <PageHeader
        title={animalTitle(a)}
        subtitle={[d.breedName, d.farmName].filter(Boolean).join(' · ')}
        actions={
          <>
            {a.isDemo && <DemoBadge />}
            <StatusBadge status={a.status} label={statusLabel[a.status]} />
            <button className="btn" onClick={() => setEditing(true)}>Editar</button>
            {archived
              ? <button className="btn" onClick={async () => {
                  try { await api.restoreAnimal(a.id); toast('success', 'Animal restaurado.'); detail.reload() } catch (e) { toast('error', errorMessage(e)) }
                }}>Restaurar</button>
              : <button className="btn btn-danger-outline" onClick={() => setArchiving(true)}>Arquivar</button>}
          </>
        }
      />
      <Tabs tabs={tabs} active={tab} onChange={setTab} label="Seções do animal" />
      <div role="tabpanel" id={`panel-${tab}`} aria-labelledby={`tab-${tab}`} className="tab-panel">
        {tab === 'summary' && <SummaryTab detail={d} />}
        {tab === 'phenotype' && <PhenotypeTab animalId={a.id} />}
        {tab === 'genetics' && <GeneticsTab animalId={a.id} />}
        {tab === 'pedigree' && <PedigreeTab detail={d} />}
        {tab === 'repro' && <ReproTab detail={d} />}
      </div>

      {editing && (
        <AnimalFormModal farms={farms} breeds={breeds} animal={a} onClose={() => setEditing(false)}
          onSaved={() => { setEditing(false); detail.reload() }} />
      )}
      {archiving && (
        <ConfirmDialog title="Arquivar animal" confirmLabel="Arquivar" danger onClose={() => setArchiving(false)}
          onConfirm={async () => { await api.archiveAnimal(a.id); toast('success', 'Animal arquivado.'); detail.reload() }}>
          <p><strong>{animalTitle(a)}</strong> deixará de aparecer na lista padrão do rebanho, mas continuará preservado no banco de dados e nos pedigrees. Você pode restaurá-lo depois.</p>
        </ConfirmDialog>
      )}
    </>
  )
}
