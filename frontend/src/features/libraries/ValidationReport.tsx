import type { ImportReport } from '../../types/domain'

const n = (v: number) => v.toLocaleString('pt-BR')

/** The pre-install summary, in the format agreed for the import wizard. */
export function ValidationReport({ report }: { report: ImportReport }) {
  return (
    <section aria-label="Validação" className="card validation">
      <h2>Validação</h2>
      <ul className="plain-list validation-list">
        <li><strong>{n(report.records)}</strong> registros</li>
        <li><strong>{n(report.traitsRecognized)}</strong> características reconhecidas</li>
        <li><strong>{n(report.withoutAccuracy)}</strong> sem acurácia</li>
        <li><strong>{n(report.duplicates)}</strong> duplicados</li>
        <li><strong>{n(report.traitsUnrecognized)}</strong> traits não reconhecidos</li>
      </ul>
      {report.traitsUnrecognized > 0 && (
        <p className="muted small">Colunas não mapeadas: {report.unrecognizedColumns.join(', ')}</p>
      )}
      <ul className="plain-list validation-list secondary">
        <li>{n(report.matchedAnimals)} correspondem a animais já cadastrados</li>
        <li>{n(report.unmatchedAnimals)} não correspondem a nenhum animal cadastrado</li>
        {report.ambiguousAnimals > 0 && <li>{n(report.ambiguousAnimals)} ambíguos (ignorados)</li>}
      </ul>
      {report.errors.length > 0 && (
        <div role="alert" className="notice notice-error">
          <strong>Problemas que impedem a importação</strong>
          <ul className="plain-list">{report.errors.map((e, i) => <li key={i}>{e}</li>)}</ul>
        </div>
      )}
      {report.warnings.length > 0 && (
        <div className="notice notice-warn">
          <strong>Avisos</strong>
          <ul className="plain-list">{report.warnings.map((w, i) => <li key={i}>{w}</li>)}</ul>
        </div>
      )}
    </section>
  )
}
