import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { PageHeader } from '../components/PageHeader'
import { chapters } from '../features/docs/chapters'
import { StatusBadge } from '../features/docs/Formula'
import { useAsync } from '../hooks/useAsync'
import { api } from '../utils/api'

type Mode = 'simple' | 'technical'
const KEY = 'genetica.docs.mode'

function loadMode(): Mode {
  try {
    return localStorage.getItem(KEY) === 'technical' ? 'technical' : 'simple'
  } catch {
    return 'simple'
  }
}

export function CalculationsPage() {
  const { section } = useParams()
  const [mode, setMode] = useState<Mode>(loadMode)
  const version = useAsync(() => api.engineVersion())

  useEffect(() => {
    try { localStorage.setItem(KEY, mode) } catch { /* storage unavailable */ }
  }, [mode])

  useEffect(() => {
    if (!section) return
    document.getElementById(section)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, [section])

  return (
    <>
      <PageHeader
        title="Cálculos"
        subtitle="Como o aplicativo chega às suas notas: o que é calculado, como e o que ainda não é"
        actions={
          <div role="group" aria-label="Nível de detalhe" className="segmented">
            <button className={mode === 'simple' ? 'seg active' : 'seg'} aria-pressed={mode === 'simple'} onClick={() => setMode('simple')}>Modo simples</button>
            <button className={mode === 'technical' ? 'seg active' : 'seg'} aria-pressed={mode === 'technical'} onClick={() => setMode('technical')}>Modo técnico</button>
          </div>
        }
      />

      <div className="doc-layout">
        <nav className="doc-toc" aria-label="Capítulos">
          <ol>
            {chapters.map((c) => (
              <li key={c.id}><Link to={`/calculos/${c.id}`} aria-current={section === c.id ? 'location' : undefined}>{c.number}. {c.title}</Link></li>
            ))}
          </ol>
          {version.data && <p className="muted small">Motor {version.data}</p>}
        </nav>

        <div className="doc-body">
          <section className="card block" aria-label="Resumo do que é calculado">
            <h2>O que este software calcula</h2>
            <p>Um ranking de touros para uma matriz é <strong>determinístico e auditável</strong>: sem IA, sem aleatoriedade, com cada número rastreável aos dados e ao método. Quando algo não pode ser calculado com os dados disponíveis, a resposta é <strong>“Não calculável com os dados disponíveis”</strong>, e isso é uma resposta válida.</p>
            <ul className="status-list">
              {chapters.map((c) => (
                <li key={c.id}><StatusBadge status={c.status} /> <Link to={`/calculos/${c.id}`}>{c.title}</Link></li>
              ))}
            </ul>
            <p className="notice notice-warn">O MVP não possui validação zootécnica. Perfis de demonstração e exemplos desta página são ilustrativos e não representam recomendação.</p>
          </section>

          {chapters.map((c) => (
            <section key={c.id} id={c.id} className="doc-chapter" aria-labelledby={`h-${c.id}`}>
              <header>
                <h2 id={`h-${c.id}`}><span className="doc-num">{c.number}</span> {c.title}</h2>
                <StatusBadge status={c.status} />
              </header>
              {c.notice && <p className="notice notice-warn doc-notice">{c.notice}</p>}
              <p className="doc-lead">{c.lead}</p>
              <div className="doc-content">{mode === 'simple' ? c.simple : c.technical}</div>
            </section>
          ))}
        </div>
      </div>
    </>
  )
}
