import { useState } from 'react'
import { Modal } from '../../components/Modal'
import { Tabs } from '../../components/Tabs'
import { useAsync } from '../../hooks/useAsync'
import { api } from '../../utils/api'
import { formatDate } from '../../utils/format'
import { directionLabel, purposeLabel, traitCategoryLabel, valueKindLabel } from '../../utils/labels'
import { LicenseBadge } from './LicenseBadge'

export type DetailSection = 'info' | 'license' | 'docs' | 'traits'

const sections = [
  { id: 'info', label: 'Detalhes' },
  { id: 'license', label: 'Licença' },
  { id: 'docs', label: 'Documentação' },
  { id: 'traits', label: 'Características' },
]

export function LibraryDetailModal({ id, initial = 'info', onClose }: { id: string; initial?: DetailSection; onClose: () => void }) {
  const [section, setSection] = useState<string>(initial)
  const lib = useAsync(() => api.getLibrary(id))

  return (
    <Modal title="Biblioteca" onClose={onClose} wide>
      {lib.loading && <p role="status" className="muted">Carregando…</p>}
      {lib.error && <div role="alert" className="notice notice-error">{lib.error}</div>}
      {lib.data && (
        <>
          <h3 className="modal-title">{lib.data.name} <span className="muted">v{lib.data.version}</span></h3>
          <Tabs tabs={sections} active={section} onChange={setSection} label="Seções da biblioteca" />
          <div role="tabpanel" id={`panel-${section}`} aria-labelledby={`tab-${section}`}>
            {section === 'info' && (
              <dl className="kv">
                <dt>Identificador</dt><dd className="mono">{lib.data.packageId}</dd>
                <dt>Origem</dt><dd>{lib.data.origin === 'genlib' ? 'Pacote .genlib' : `Arquivo ${lib.data.origin.toUpperCase()} importado`}</dd>
                <dt>Espécie</dt><dd>{lib.data.species || '—'}</dd>
                <dt>Raças</dt><dd>{lib.data.breeds.join(', ') || '—'}</dd>
                <dt>Finalidade</dt><dd>{lib.data.purposes.map((p) => purposeLabel[p] ?? p).join(', ') || '—'}</dd>
                <dt>Organização</dt><dd>{lib.data.organization || '—'}</dd>
                <dt>Fonte</dt><dd className="mono">{lib.data.sourceUrl || '—'}</dd>
                <dt>Programa</dt><dd>{lib.data.program}</dd>
                <dt>Release</dt><dd>{lib.data.release}</dd>
                <dt>Registros</dt><dd>{lib.data.recordCount.toLocaleString('pt-BR')}</dd>
                <dt>Características</dt><dd>{lib.data.traitCount}</dd>
                <dt>Instalada em</dt><dd>{formatDate(lib.data.installedAt)}</dd>
                <dt>Usada em cálculos</dt><dd>{lib.data.usedByCalculations}</dd>
                <dt>SHA-256</dt><dd className="mono">{lib.data.fileSha256}</dd>
              </dl>
            )}
            {section === 'license' && (
              <>
                <p><LicenseBadge status={lib.data.licenseStatus} name={lib.data.licenseName} /> {lib.data.licenseName && <strong>{lib.data.licenseName}</strong>}</p>
                {lib.data.licenseUrl && <p className="mono">{lib.data.licenseUrl}</p>}
                <p className="notice notice-info">{lib.data.responsibilityNotice}</p>
                <pre className="doc-text">{lib.data.licenseText || 'Nenhum texto de licença acompanha este pacote.'}</pre>
              </>
            )}
            {section === 'docs' && <pre className="doc-text">{lib.data.readme || 'Este pacote não inclui documentação.'}</pre>}
            {section === 'traits' && (
              <table className="data-table compact">
                <thead><tr><th>Código</th><th>Nome</th><th>Categoria</th><th>Tipo</th><th>Unidade</th><th>Direção</th></tr></thead>
                <tbody>
                  {lib.data.traits.map((t) => (
                    <tr key={t.id}>
                      <td className="mono">{t.code}</td><td>{t.name}</td><td>{traitCategoryLabel[t.category]}</td>
                      <td>{valueKindLabel[t.evaluationType]}</td><td>{t.unit || '—'}</td><td>{directionLabel[t.optimizationDirection]}</td>
                    </tr>
                  ))}
                  {lib.data.traits.length === 0 && <tr><td colSpan={6} className="muted">Usa apenas características do catálogo.</td></tr>}
                </tbody>
              </table>
            )}
          </div>
        </>
      )}
    </Modal>
  )
}
