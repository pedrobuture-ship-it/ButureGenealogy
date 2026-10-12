import { useState } from 'react'
import type { IntegrityReport } from '../../types/domain'
import { api, errorMessage } from '../../utils/api'
import { formatDateTime } from '../../utils/format'

export function IntegritySection() {
  const [report, setReport] = useState<IntegrityReport | null>(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function run() {
    setBusy(true); setError(null)
    try {
      setReport(await api.runIntegrityCheck())
    } catch (e) {
      setError(errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  const errors = report?.issues.filter((i) => i.severity === 'error') ?? []
  const warnings = report?.issues.filter((i) => i.severity === 'warning') ?? []
  return (
    <section className="card block" aria-labelledby="sec-integrity">
      <h2 id="sec-integrity">Integridade dos dados</h2>
      <p className="muted">Verifica o arquivo do banco, as migrations, as referências entre registros, os pacotes de bibliotecas, os cálculos históricos e o estoque. A verificação apenas informa: nunca altera nada.</p>
      <button className="btn btn-primary" onClick={run} disabled={busy}>{busy ? 'Verificando…' : 'Verificar integridade'}</button>
      {error && <div role="alert" className="notice notice-error">{error}</div>}
      {report && (
        <div className="spaced" aria-live="polite">
          {report.ok && report.warnings === 0 ? (
            <div role="status" className="notice notice-info"><strong>Nenhum problema encontrado.</strong><span>Verificado em {formatDateTime(report.checkedAt)}.</span></div>
          ) : (
            <p><strong>{report.errors}</strong> erro(s) · <strong>{report.warnings}</strong> aviso(s) · verificado em {formatDateTime(report.checkedAt)}</p>
          )}
          {errors.length > 0 && (
            <div role="alert" className="notice notice-error">
              <strong>Erros</strong>
              <ul className="plain-list">{errors.map((i, k) => <li key={k}>{i.message}{i.detail && <span className="muted small"> — {i.detail}</span>}</li>)}</ul>
            </div>
          )}
          {warnings.length > 0 && (
            <div className="notice notice-warn">
              <strong>Avisos</strong>
              <ul className="plain-list">{warnings.map((i, k) => <li key={k}>{i.message}{i.detail && <span className="small"> — {i.detail}</span>}</li>)}</ul>
            </div>
          )}
        </div>
      )}
    </section>
  )
}
