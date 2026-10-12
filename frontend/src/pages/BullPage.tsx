import { useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { DemoBadge, StatusBadge } from '../components/Badge'
import { PageHeader } from '../components/PageHeader'
import { Tabs } from '../components/Tabs'
import { GeneticsTab } from '../features/animal/GeneticsTab'
import { AnimalFormModal } from '../features/herd/AnimalFormModal'
import { PedigreeView } from '../features/pedigree/PedigreeView'
import { MovementsTab } from '../features/semen/MovementsTab'
import { StockBadge } from '../features/semen/StockBadge'
import { StockTab } from '../features/semen/StockTab'
import { useAsync } from '../hooks/useAsync'
import { api } from '../utils/api'
import { animalTitle } from '../utils/format'
import { statusLabel } from '../utils/labels'

const tabs = [
  { id: 'stock', label: 'Estoque' },
  { id: 'genetics', label: 'Genética' },
  { id: 'pedigree', label: 'Genealogia' },
  { id: 'movements', label: 'Movimentações' },
]

export function BullPage() {
  const { id = '' } = useParams()
  const [params, setParams] = useSearchParams()
  const [tab, setTab] = useState('stock')
  const [editing, setEditing] = useState(false)
  const [openNewLot, setOpenNewLot] = useState(params.get('novoLote') === '1')

  const data = useAsync(async () => {
    const [sheet, types, farms, breeds] = await Promise.all([api.getBullSheet(id), api.listSemenTypes(), api.listFarms(), api.listBreeds()])
    return { sheet, types, farms, breeds }
  })
  const [lastId, setLastId] = useState(id)
  if (lastId !== id) { setLastId(id); data.reload() }

  if (data.loading && !data.data) return <p role="status" className="muted">Carregando touro…</p>
  if (data.error || !data.data) {
    return (<><div role="alert" className="notice notice-error">{data.error ?? 'Touro não encontrado.'}</div><Link to="/semen">Voltar ao banco de sêmen</Link></>)
  }
  const { sheet, types, farms, breeds } = data.data
  const a = sheet.detail.animal
  const typeName = (code: string) => types.find((t) => t.code === code)?.name ?? code
  const clearParam = () => { if (params.has('novoLote')) { params.delete('novoLote'); setParams(params, { replace: true }) } }

  return (
    <>
      <p className="breadcrumb"><Link to="/semen">Sêmen</Link> / {animalTitle(a)}</p>
      <PageHeader title={animalTitle(a)}
        subtitle={[sheet.detail.breedName, a.registration && `Reg. ${a.registration}`, sheet.detail.farmName].filter(Boolean).join(' · ')}
        actions={<>
          {a.isDemo && <DemoBadge />}
          {a.status !== 'active' && <StatusBadge status={a.status} label={statusLabel[a.status]} />}
          <StockBadge status={sheet.stockStatus} />
          <button className="btn" onClick={() => setEditing(true)}>Editar touro</button>
          <Link className="btn" to={`/rebanho/${a.id}`}>Ficha completa</Link>
        </>} />
      <Tabs tabs={tabs} active={tab} onChange={setTab} label="Seções do touro" />
      <div role="tabpanel" id={`panel-${tab}`} aria-labelledby={`tab-${tab}`} className="tab-panel">
        {tab === 'stock' && <StockTab sheet={sheet} typeName={typeName} openNewLot={openNewLot} onChanged={() => { setOpenNewLot(false); clearParam(); data.reload() }} />}
        {tab === 'genetics' && <GeneticsTab animalId={a.id} />}
        {tab === 'pedigree' && <PedigreeView animalId={a.id} />}
        {tab === 'movements' && <MovementsTab sheet={sheet} />}
      </div>
      {editing && <AnimalFormModal farms={farms} breeds={breeds} animal={a} onClose={() => setEditing(false)} onSaved={() => { setEditing(false); data.reload() }} />}
    </>
  )
}
