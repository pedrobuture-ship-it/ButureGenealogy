import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { Icon } from '../components/Icon'
import { navItems } from './navigation'

function isEditable(el: EventTarget | null): boolean {
  if (!(el instanceof HTMLElement)) return false
  return el.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName)
}

/** Alt+1…9 navigates the sidebar; "/" focuses the page's search box. */
function useGlobalShortcuts() {
  const navigate = useNavigate()
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.altKey && !e.ctrlKey && !e.metaKey && /^[1-9]$/.test(e.key)) {
        const item = navItems[Number(e.key) - 1]
        if (item) { e.preventDefault(); navigate(item.path) }
        return
      }
      if (e.key === '/' && !e.ctrlKey && !e.metaKey && !e.altKey && !isEditable(e.target)) {
        const box = document.querySelector<HTMLInputElement>('main input[type="search"]')
        if (box) { e.preventDefault(); box.focus() }
      }
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [navigate])
}

export function AppLayout() {
  useGlobalShortcuts()
  const location = useLocation()
  return (
    <div className="shell">
      <a className="skip-link" href="#main">Ir para o conteúdo</a>
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-mark" aria-hidden="true" />
          <div>
            <div className="brand-name">Genética</div>
            <div className="brand-sub">Melhoramento bovino</div>
          </div>
        </div>
        <nav aria-label="Navegação principal">
          <ul className="nav-list">
            {navItems.map((item, i) => (
              <li key={item.path}>
                <NavLink to={item.path} end={item.path === '/'} className="nav-link" title={`${item.label} (Alt+${i + 1})`}>
                  <Icon name={item.icon} />
                  <span>{item.label}</span>
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>
        <div className="sidebar-foot">Dados armazenados localmente</div>
      </aside>
      <main id="main" className="content" tabIndex={-1}>
        <ErrorBoundary resetKey={location.pathname}><Outlet /></ErrorBoundary>
      </main>
    </div>
  )
}
