import { navItems } from '../../layouts/navigation'

export function ShortcutsSection() {
  return (
    <section className="card block" aria-labelledby="sec-shortcuts">
      <h2 id="sec-shortcuts">Atalhos de teclado</h2>
      <table className="data-table compact">
        <caption className="sr-only">Atalhos de teclado</caption>
        <thead><tr><th>Atalho</th><th>Ação</th></tr></thead>
        <tbody>
          {navItems.map((n, i) => (<tr key={n.path}><td><kbd>Alt</kbd> + <kbd>{i + 1}</kbd></td><td>Ir para {n.label}</td></tr>))}
          <tr><td><kbd>/</kbd></td><td>Focar a busca da página</td></tr>
          <tr><td><kbd>Esc</kbd></td><td>Fechar janelas e painéis</td></tr>
          <tr><td><kbd>←</kbd> <kbd>→</kbd></td><td>Trocar de aba (com o foco nas abas)</td></tr>
        </tbody>
      </table>
    </section>
  )
}
