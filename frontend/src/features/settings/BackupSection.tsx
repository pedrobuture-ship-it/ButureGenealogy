import { useState } from 'react'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { useToast } from '../../components/Toast'
import { useAsync } from '../../hooks/useAsync'
import type { BackupManifest, BackupPreview, RestoreResult } from '../../types/domain'
import { api, errorMessage } from '../../utils/api'
import { formatDateTime } from '../../utils/format'

const countLabels: [string, string][] = [['farms', 'propriedades'], ['animals', 'animais'], ['animal_measurements', 'medições'], ['genetic_evaluations', 'avaliações genéticas'],
  ['semen_inventory', 'lotes de sêmen'], ['matings', 'cálculos salvos'], ['datasets', 'bibliotecas']]

export function BackupSection() {
  const toast = useToast()
  const auto = useAsync(() => api.listAutomaticBackups())
  const [pathInput, setPathInput] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [created, setCreated] = useState<{ path: string; manifest: BackupManifest } | null>(null)
  const [preview, setPreview] = useState<BackupPreview | null>(null)
  const [restored, setRestored] = useState<RestoreResult | null>(null)

  async function create() {
    setBusy(true); setError(null); setCreated(null)
    try {
      const suggested = await api.suggestedBackupName()
      const path = await api.pickBackupSavePath(suggested)
      if (!path) return
      const manifest = await api.createBackup(path)
      setCreated({ path, manifest })
      toast('success', 'Backup criado.')
    } catch (e) {
      setError(errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  async function inspect(path: string) {
    if (!path.trim()) return
    setBusy(true); setError(null); setRestored(null)
    try {
      setPreview(await api.inspectBackup(path.trim()))
    } catch (e) {
      setError(errorMessage(e)); setPreview(null)
    } finally {
      setBusy(false)
    }
  }

  async function pickAndInspect() {
    setError(null)
    try {
      const p = await api.pickBackupOpenPath()
      if (p) { setPathInput(p); await inspect(p) }
    } catch (e) {
      setError(errorMessage(e))
    }
  }

  return (
    <section className="card block" aria-labelledby="sec-backup">
      <h2 id="sec-backup">Backup e restauração</h2>
      <p className="muted">O backup (<span className="mono">.genbackup</span>) reúne o banco de dados, bibliotecas particulares, configurações e anexos, com verificação de integridade (SHA-256). Bibliotecas públicas entram só como metadados de versão.</p>
      {error && <div role="alert" className="notice notice-error">{error}</div>}

      <div className="row-gap">
        <button className="btn btn-primary" onClick={create} disabled={busy}>Criar backup</button>
      </div>
      {created && (
        <div role="status" className="notice notice-info">
          <strong>Backup criado</strong>
          <span className="mono">{created.path}</span>
          <span>{created.manifest.counts.animals ?? 0} animais · {created.manifest.libraries.length} biblioteca(s) · versão do aplicativo {created.manifest.appVersion}</span>
        </div>
      )}

      <h3 className="sub-h">Restaurar</h3>
      <div className="row-gap">
        <button className="btn" onClick={pickAndInspect} disabled={busy}>Restaurar backup…</button>
        <span className="muted">ou informe o caminho:</span>
        <input aria-label="Caminho do backup" className="grow" value={pathInput} onChange={(e) => setPathInput(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') void inspect(pathInput) }} placeholder="/caminho/arquivo.genbackup" />
        <button className="btn" onClick={() => inspect(pathInput)} disabled={busy || !pathInput.trim()}>Verificar</button>
      </div>
      <p className="muted small">Antes de restaurar, o arquivo é validado (formato, integridade e versão) e um backup de segurança da base atual é criado automaticamente. Nada é sobrescrito sem a sua confirmação.</p>

      {restored && (
        <div role="status" className="notice notice-info">
          <strong>Backup restaurado</strong>
          <span>Backup de segurança da base anterior: <span className="mono">{restored.safetyBackup}</span></span>
          {restored.missingLibraries.length > 0 && <span>Bibliotecas públicas sem o pacote original (apenas metadados): {restored.missingLibraries.join(', ')}. Os valores estão no banco.</span>}
        </div>
      )}

      {auto.data && auto.data.length > 0 && (
        <details>
          <summary>Backups de segurança automáticos ({auto.data.length})</summary>
          <ul className="plain-list mono small">{auto.data.map((p) => <li key={p}>{p}</li>)}</ul>
        </details>
      )}

      {preview && <RestoreDialog preview={preview} onClose={() => setPreview(null)} onDone={(r) => { setRestored(r); setPreview(null); auto.reload(); toast('success', 'Backup restaurado.') }} />}
    </section>
  )
}

function RestoreDialog({ preview, onClose, onDone }: { preview: BackupPreview; onClose: () => void; onDone: (r: RestoreResult) => void }) {
  const [understood, setUnderstood] = useState(false)
  const m = preview.manifest
  const upgraded = preview.upgradedFrom < preview.schemaVersion
  return (
    <ConfirmDialog title="Restaurar backup" confirmLabel="Substituir os dados atuais" danger onClose={onClose}
      onConfirm={async () => {
        if (!understood) throw new Error('Marque a confirmação para continuar.')
        onDone(await api.restoreBackup(preview.path))
      }}>
      <p>O arquivo foi validado: formato, integridade (SHA-256 e SQLite) e versão estão corretos.</p>
      <dl className="kv">
        <dt>Criado em</dt><dd>{formatDateTime(m.createdAt)}</dd>
        <dt>Versão do aplicativo</dt><dd>{m.appVersion || '—'}</dd>
        <dt>Propriedades</dt><dd>{m.farms.join(', ') || '—'}</dd>
        {countLabels.filter(([k]) => m.counts[k] != null).map(([k, l]) => (<span key={k} className="kv-row"><dt>{l[0].toUpperCase() + l.slice(1)}</dt><dd>{m.counts[k]}</dd></span>))}
        <dt>Bibliotecas</dt><dd>{m.libraries.length === 0 ? '—' : m.libraries.map((l) => `${l.name} ${l.version}${l.included ? '' : ' (somente metadados)'}`).join('; ')}</dd>
      </dl>
      {upgraded && <p className="notice notice-info">Este backup é de uma versão anterior (esquema {preview.upgradedFrom}) e será atualizado para a versão atual ({preview.schemaVersion}).</p>}
      <p className="notice notice-warn"><strong>Os dados atuais serão substituídos.</strong> Um backup de segurança da base atual será criado automaticamente antes.</p>
      <label className="check"><input type="checkbox" checked={understood} onChange={(e) => setUnderstood(e.target.checked)} /> Entendo que os dados atuais serão substituídos por este backup.</label>
    </ConfirmDialog>
  )
}
