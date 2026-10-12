import { HashRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AppLayout } from './layouts/AppLayout'
import { navItems } from './layouts/navigation'
import { ToastProvider } from './components/Toast'
import { AnimalDetailPage } from './pages/AnimalDetailPage'
import { BullPage } from './pages/BullPage'
import { SemenPage } from './pages/SemenPage'
import { SettingsPage } from './pages/SettingsPage'
import { CalculationsPage } from './pages/CalculationsPage'
import { MatingPage } from './pages/MatingPage'
import { LibrariesPage } from './pages/LibrariesPage'
import { PedigreePage } from './pages/PedigreePage'
import { HerdPage } from './pages/HerdPage'
import { HomePage } from './pages/HomePage'
import { Placeholder } from './pages/Placeholder'

export default function App() {
  return (
    <ToastProvider>
    <HashRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route index element={<HomePage />} />
          <Route path="/rebanho" element={<HerdPage />} />
          <Route path="/rebanho/:id" element={<AnimalDetailPage />} />
          <Route path="/genealogia/:id?" element={<PedigreePage />} />
          <Route path="/bibliotecas" element={<LibrariesPage />} />
          <Route path="/semen" element={<SemenPage />} />
          <Route path="/acasalamentos" element={<MatingPage />} />
          <Route path="/calculos/:section?" element={<CalculationsPage />} />
          <Route path="/configuracoes" element={<SettingsPage />} />
          <Route path="/semen/:id" element={<BullPage />} />
          {navItems
            .filter((n) => n.path !== '/' && n.path !== '/rebanho' && n.path !== '/genealogia' && n.path !== '/bibliotecas' && n.path !== '/semen' && n.path !== '/acasalamentos' && n.path !== '/calculos' && n.path !== '/configuracoes')
            .map((n) => (
              <Route key={n.path} path={n.path} element={<Placeholder path={n.path} />} />
            ))}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </HashRouter>
    </ToastProvider>
  )
}
