import { useState } from 'react'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { useToast } from '../../components/Toast'
import { useAsync } from '../../hooks/useAsync'
import { api, errorMessage } from '../../utils/api'

export function DemoSection() {
  const toast = useToast()
  const has = useAsync(() => api.hasDemoData())
  const [busy, setBusy] = useState(false)
  const [removing, setRemoving] = useState(false)

  async function load() {
    setBusy(true)
    try {
      await api.loadDemoData()
      toast('success', 'Dados DEMO / SYNTHETIC carregados.')
      has.reload()
    } catch (e) {
      toast('error', errorMessage(e))
    } finally {
      setBusy(false)
    }
  }

  return (
    <section className="card block" aria-labelledby="sec-demo">
      <h2 id="sec-demo">Dados de demonstração</h2>
      <p className="muted">Animais, avaliações e cálculos <strong>fictícios (DEMO / SYNTHETIC)</strong> para conhecer o aplicativo. Não representam dados científicos reais e são sempre identificados na interface.</p>
      {has.error && <div role="alert" className="notice notice-error">{has.error}</div>}
      {has.data === false && <button className="btn" onClick={load} disabled={busy}>{busy ? 'Carregando…' : 'Carregar dados de demonstração'}</button>}
      {has.data === true && <button className="btn btn-danger-outline" onClick={() => setRemoving(true)}>Remover dados de demonstração</button>}
      {removing && (
        <ConfirmDialog title="Remover dados de demonstração" confirmLabel="Remover" danger onClose={() => setRemoving(false)}
          onConfirm={async () => { await api.removeDemoData(); toast('success', 'Dados de demonstração removidos.'); has.reload() }}>
          Remove somente os registros marcados como DEMO / SYNTHETIC (animais, avaliações, bibliotecas e cálculos de demonstração). Seus dados reais não são afetados.
        </ConfirmDialog>
      )}
    </section>
  )
}
