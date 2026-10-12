import { PageHeader } from '../components/PageHeader'
import { useHealth } from '../features/health/useHealth'

export function HomePage() {
  const { data, error, loading, reload } = useHealth()

  return (
    <>
      <PageHeader
        title="Início"
        subtitle="Visão geral do sistema"
        actions={<button className="btn" onClick={reload}>Verificar novamente</button>}
      />
      {loading && <p role="status" className="muted">Verificando o banco de dados…</p>}
      {error && (
        <div role="alert" className="notice notice-error">
          <strong>Não foi possível consultar o sistema.</strong>
          <span>{error}</span>
        </div>
      )}
      {data && (
        <section aria-label="Estado do sistema" className="grid">
          <div className="card">
            <h2>Banco de dados</h2>
            <p className={data.db.ok ? 'status ok' : 'status bad'}>{data.db.ok ? 'Saudável' : 'Com problemas'}</p>
            <dl className="kv">
              <dt>Versão do esquema</dt><dd>{data.db.schemaVersion}</dd>
              <dt>Journal</dt><dd>{data.db.journalMode}</dd>
              <dt>Chaves estrangeiras</dt><dd>{data.db.foreignKeys ? 'ativas' : 'inativas'}</dd>
              <dt>Verificação</dt><dd>{data.db.quickCheck}</dd>
            </dl>
            {data.db.error && <p className="notice notice-error">{data.db.error}</p>}
          </div>
          <div className="card">
            <h2>Aplicativo</h2>
            <dl className="kv">
              <dt>Versão</dt><dd>{data.info.version}</dd>
              <dt>Plataforma</dt><dd>{data.info.platform}</dd>
              <dt>Dados em</dt><dd className="mono">{data.info.dataDir}</dd>
            </dl>
          </div>
        </section>
      )}
    </>
  )
}
