import { useState } from 'react'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { EmptyState } from '../../components/EmptyState'
import { DemoBadge } from '../../components/Badge'
import { useToast } from '../../components/Toast'
import { useAsync } from '../../hooks/useAsync'
import type { LibrarySummary } from '../../types/domain'
import { api, errorMessage } from '../../utils/api'
import { purposeLabel } from '../../utils/labels'
import { LibraryDetailModal, type DetailSection } from './LibraryDetailModal'
import { LicenseBadge } from './LicenseBadge'

export function InstalledTab({ onImport }: { onImport: () => void }) {
  const toast = useToast()
  const libs = useAsync(() => api.listLibraries())
  const [detail, setDetail] = useState<{ id: string; section: DetailSection } | null>(null)
  const [removing, setRemoving] = useState<LibrarySummary | null>(null)

  async function toggle(l: LibrarySummary) {
    try {
      await api.setLibraryActive(l.id, !l.active)
      toast('success', l.active ? 'Biblioteca desativada.' : 'Biblioteca ativada.')
      libs.reload()
    } catch (e) {
      toast('error', errorMessage(e))
    }
  }

  if (libs.loading && !libs.data) return <p role="status" className="muted">Carregando bibliotecas…</p>
  if (libs.error) return <div role="alert" className="notice notice-error">{libs.error}</div>
  const list = libs.data ?? []

  if (list.length === 0) {
    return (
      <EmptyState title="Nenhuma biblioteca instalada">
        Bibliotecas trazem avaliações genéticas e definições de características. Importe um pacote .genlib ou um arquivo CSV/JSON.
        <span className="empty-actions"><button className="btn btn-primary" onClick={onImport}>Importar biblioteca</button></span>
      </EmptyState>
    )
  }

  return (
    <>
      <div className="table-wrap">
        <table className="data-table wrap">
          <caption className="sr-only">Bibliotecas instaladas</caption>
          <thead>
            <tr>
              <th>Nome</th><th>Versão</th><th>Raça</th><th>Finalidade</th><th>Fonte</th><th>Status</th><th>Licença</th><th>Registros</th><th>Traits</th><th><span className="sr-only">Ações</span></th>
            </tr>
          </thead>
          <tbody>
            {list.map((l) => (
              <tr key={l.id} className={l.active ? '' : 'row-inactive'}>
                <td><strong>{l.name}</strong> {l.isDemo && <DemoBadge />}</td>
                <td className="mono nowrap">{l.version}</td>
                <td>{l.breeds.join(', ') || '—'}</td>
                <td>{l.purposes.map((p) => purposeLabel[p] ?? p).join(', ') || '—'}</td>
                <td>{l.organization || (l.origin === 'genlib' ? 'Pacote .genlib' : `Arquivo ${l.origin.toUpperCase()}`)}</td>
                <td><span className={l.active ? 'badge badge-ok' : 'badge badge-muted'}>{l.active ? 'Ativa' : 'Inativa'}</span></td>
                <td><LicenseBadge status={l.licenseStatus} name={l.licenseName} /></td>
                <td>≈ {l.recordCount.toLocaleString('pt-BR')}</td>
                <td>{l.traitCount}</td>
                <td className="actions-cell">
                  <button className="btn btn-link" onClick={() => setDetail({ id: l.id, section: 'info' })}>Detalhes</button>
                  <button className="btn btn-link" onClick={() => setDetail({ id: l.id, section: 'license' })}>Licença</button>
                  <button className="btn btn-link" onClick={() => setDetail({ id: l.id, section: 'docs' })}>Documentação</button>
                  <button className="btn btn-link" onClick={() => toggle(l)}>{l.active ? 'Desativar' : 'Ativar'}</button>
                  <button className="btn btn-link danger-link" onClick={() => setRemoving(l)}>Remover</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {detail && <LibraryDetailModal id={detail.id} initial={detail.section} onClose={() => setDetail(null)} />}
      {removing && (
        <ConfirmDialog title="Remover biblioteca" confirmLabel="Remover" danger onClose={() => setRemoving(null)}
          onConfirm={async () => { await api.removeLibrary(removing.id); toast('success', 'Biblioteca removida.'); libs.reload() }}>
          <p>Remover <strong>{removing.name} v{removing.version}</strong> apaga as avaliações importadas por ela e as características que ela definiu. Os animais cadastrados continuam no rebanho.</p>
          <p className="muted">Versões usadas em cálculos salvos não podem ser removidas (use “Desativar”).</p>
        </ConfirmDialog>
      )}
    </>
  )
}
