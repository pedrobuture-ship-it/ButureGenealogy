import { PageHeader } from '../components/PageHeader'
import { BackupSection } from '../features/settings/BackupSection'
import { DemoSection } from '../features/settings/DemoSection'
import { IntegritySection } from '../features/settings/IntegritySection'
import { ShortcutsSection } from '../features/settings/ShortcutsSection'
import { useHealth } from '../features/health/useHealth'
import { useAsync } from '../hooks/useAsync'
import { api } from '../utils/api'

export function SettingsPage() {
  const health = useHealth()
  const engine = useAsync(() => api.engineVersion())
  return (
    <>
      <PageHeader title="Configurações" subtitle="Backup, integridade, dados de demonstração e preferências" />
      <BackupSection />
      <IntegritySection />
      <DemoSection />
      <ShortcutsSection />
      <section className="card block" aria-labelledby="sec-about">
        <h2 id="sec-about">Sobre</h2>
        {health.data ? (
          <dl className="kv">
            <dt>Aplicativo</dt><dd>{health.data.info.name} {health.data.info.version}</dd>
            <dt>Motor de acasalamento</dt><dd>{engine.data ?? '—'}</dd>
            <dt>Plataforma</dt><dd>{health.data.info.platform}</dd>
            <dt>Dados em</dt><dd className="mono">{health.data.info.dataDir}</dd>
            <dt>Banco de dados</dt><dd>{health.data.db.ok ? 'saudável' : 'com problemas'} · esquema {health.data.db.schemaVersion}</dd>
          </dl>
        ) : <p className="muted">Carregando…</p>}
        <p className="muted small">Funciona totalmente offline. Seus dados ficam apenas neste computador.</p>
      </section>
    </>
  )
}
