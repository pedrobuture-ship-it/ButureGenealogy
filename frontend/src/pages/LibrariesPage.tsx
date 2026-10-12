import { useState } from 'react'
import { PageHeader } from '../components/PageHeader'
import { Tabs } from '../components/Tabs'
import { CatalogTab } from '../features/libraries/CatalogTab'
import { ImportWizard } from '../features/libraries/ImportWizard'
import { InstalledTab } from '../features/libraries/InstalledTab'

const tabs = [
  { id: 'installed', label: 'Instaladas' },
  { id: 'import', label: 'Importar biblioteca' },
  { id: 'catalogue', label: 'Características' },
]

export function LibrariesPage() {
  const [tab, setTab] = useState('installed')
  // remount the installed list after an install so it refetches
  const [version, setVersion] = useState(0)

  return (
    <>
      <PageHeader title="Bibliotecas" subtitle="Datasets científicos instaláveis: avaliações e definições de características" />
      <Tabs tabs={tabs} active={tab} onChange={setTab} label="Seções de bibliotecas" />
      <div role="tabpanel" id={`panel-${tab}`} aria-labelledby={`tab-${tab}`} className="tab-panel">
        {tab === 'installed' && <InstalledTab key={version} onImport={() => setTab('import')} />}
        {tab === 'import' && <ImportWizard onInstalled={() => { setVersion((v) => v + 1); setTab('installed') }} />}
        {tab === 'catalogue' && <CatalogTab />}
      </div>
    </>
  )
}
