import { fireEvent, render, screen } from '@testing-library/react'
import App from './App'
import { installFakeBackend } from './test/fakeBackend'

describe('navegação', () => {
  beforeEach(() => {
    window.location.hash = '#/'
    installFakeBackend()
  })

  it('exibe todos os itens da sidebar', async () => {
    render(<App />)
    await screen.findByText('Saudável')
    for (const label of ['Início', 'Rebanho', 'Genealogia', 'Sêmen', 'Acasalamentos', 'Bibliotecas', 'Cálculos', 'Relatórios', 'Configurações']) {
      expect(screen.getByRole('link', { name: label })).toBeInTheDocument()
    }
  })

  it('navega entre as páginas', async () => {
    render(<App />)
    await screen.findByText('Saudável')
    fireEvent.click(screen.getByRole('link', { name: 'Rebanho' }))
    expect(await screen.findByRole('heading', { level: 1, name: 'Rebanho' })).toBeInTheDocument()
    fireEvent.click(screen.getByRole('link', { name: 'Cálculos' }))
    expect(await screen.findByRole('heading', { level: 1, name: 'Cálculos' })).toBeInTheDocument()
  })

  it('mostra saúde do banco na página inicial', async () => {
    render(<App />)
    expect(await screen.findByText('Saudável')).toBeInTheDocument()
  })
})
