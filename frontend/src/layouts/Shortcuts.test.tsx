import { fireEvent, render, screen } from '@testing-library/react'
import { Component, type ReactNode } from 'react'
import App from '../App'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { ToastProvider } from '../components/Toast'
import { installFakeBackend, sampleFarm } from '../test/fakeBackend'

describe('atalhos de teclado', () => {
  it('Alt+número navega pela sidebar', async () => {
    installFakeBackend({ farms: [sampleFarm] })
    window.location.hash = '#/'
    render(<App />)
    await screen.findByText('Saudável')
    fireEvent.keyDown(document, { key: '2', altKey: true })
    expect(await screen.findByRole('heading', { level: 1, name: 'Rebanho' })).toBeInTheDocument()
    fireEvent.keyDown(document, { key: '7', altKey: true })
    expect(await screen.findByRole('heading', { level: 1, name: 'Cálculos' })).toBeInTheDocument()
  })

  it('"/" foca a busca, mas não quando já se está digitando', async () => {
    installFakeBackend({ farms: [sampleFarm] })
    window.location.hash = '#/rebanho'
    render(<App />)
    const box = await screen.findByLabelText('Buscar por nome ou registro')
    fireEvent.keyDown(document.body, { key: '/' })
    expect(box).toHaveFocus()
    fireEvent.change(box, { target: { value: 'a' } })
    fireEvent.keyDown(box, { key: '/' }) // typing inside the box: no hijack
    expect(box).toHaveFocus()
  })
})

class Bomb extends Component<{ children?: ReactNode }> {
  render(): ReactNode { throw new Error('falha de teste') }
}

describe('tratamento de erros', () => {
  it('uma falha ao renderizar mostra uma mensagem em vez de uma tela em branco', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(<ErrorBoundary><Bomb /></ErrorBoundary>)
    expect(screen.getByRole('alert')).toHaveTextContent('Algo deu errado ao exibir esta página.')
    expect(screen.getByRole('alert')).toHaveTextContent('Seus dados não foram afetados')
    expect(screen.getByRole('button', { name: 'Recarregar' })).toBeInTheDocument()
    spy.mockRestore()
  })

  it('uma promessa rejeitada sem tratamento aparece como aviso ao usuário', async () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(<ToastProvider><p>x</p></ToastProvider>)
    const ev = new Event('unhandledrejection') as Event & { reason: unknown }
    ev.reason = new Error('rede indisponível')
    window.dispatchEvent(ev)
    expect(await screen.findByText('rede indisponível')).toBeInTheDocument()
    spy.mockRestore()
  })
})
