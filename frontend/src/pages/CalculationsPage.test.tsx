import { fireEvent, render, screen, within } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { chapters } from '../features/docs/chapters'
import { installFakeBackend } from '../test/fakeBackend'
import { CalculationsPage } from './CalculationsPage'

const renderAt = (path: string) => {
  installFakeBackend()
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes><Route path="/calculos/:section?" element={<CalculationsPage />} /></Routes>
    </MemoryRouter>,
  )
}

describe('Cálculos', () => {
  beforeEach(() => localStorage.clear())

  it('tem os 13 capítulos, na ordem combinada', () => {
    renderAt('/calculos')
    const titles = chapters.map((c) => c.title)
    expect(titles).toEqual(['Score de acasalamento', 'Pesos', 'Normalização', 'MAXIMIZE', 'MINIMIZE', 'TARGET_RANGE', 'Acurácia', 'Complementaridade', 'Pedigree', 'Parentesco / endogamia', 'Progênie esperada', 'Datasets', 'Reprodutibilidade'])
    const toc = within(screen.getByRole('navigation', { name: 'Capítulos' }))
    expect(toc.getAllByRole('link')).toHaveLength(13)
    for (const [i, c] of chapters.entries()) {
      expect(document.getElementById(c.id)).toBeInTheDocument()
      expect(toc.getAllByRole('link')[i]).toHaveAttribute('href', `/calculos/${c.id}`)
    }
  })

  it('o modo simples é o padrão e o modo técnico mostra fórmulas e parâmetros, lembrando a escolha', () => {
    const { unmount } = renderAt('/calculos')
    const norm = () => within(document.getElementById('normalizacao')!)
    expect(norm().queryByText(/percentileSemantics/)).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Modo simples' })).toHaveAttribute('aria-pressed', 'true')

    fireEvent.click(screen.getByRole('button', { name: 'Modo técnico' }))
    expect(norm().getAllByText(/percentileSemantics/).length).toBeGreaterThan(0)
    expect(norm().getByLabelText('Escore z')).toHaveTextContent('Φ(z)')
    expect(within(document.getElementById('complementaridade')!).getByLabelText('Fórmula')).toHaveTextContent('bônus_máximo')
    unmount()

    renderAt('/calculos') // persisted
    expect(screen.getByRole('button', { name: 'Modo técnico' })).toHaveAttribute('aria-pressed', 'true')
  })

  it('rola até a seção da URL (links do "Ver cálculo")', () => {
    const spy = vi.spyOn(window.HTMLElement.prototype, 'scrollIntoView')
    renderAt('/calculos/parentesco')
    expect(spy).toHaveBeenCalled()
    expect((spy.mock.contexts[spy.mock.contexts.length - 1] as HTMLElement).id).toBe('parentesco')
    spy.mockRestore()
  })

  it('parentesco/endogamia é "Planejado — ainda não utilizado no score do MVP" nos dois modos, sem fórmula simulada', () => {
    renderAt('/calculos')
    const ch = () => within(document.getElementById('parentesco')!)
    expect(ch().getAllByText(/Planejado — ainda não utilizado no score do MVP/).length).toBeGreaterThan(0)
    expect(ch().getByText('Planejado', { selector: '.badge' })).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Modo técnico' }))
    expect(ch().getAllByText(/Planejado — ainda não utilizado no score do MVP/).length).toBeGreaterThan(0)
    expect(ch().getByText(/Não está implementado/i)).toBeInTheDocument()
    expect(ch().queryByLabelText('Fórmula')).not.toBeInTheDocument()
  })

  it('progênie esperada declara que não está implementada e que DEP, EBV e PTA não são a mesma escala', () => {
    renderAt('/calculos')
    const ch = within(document.getElementById('progenie')!)
    expect(ch.getByText('Não implementado', { selector: '.badge' })).toBeInTheDocument()
    expect(ch.getByText(/não devem ser tratados como se fossem a mesma escala/)).toBeInTheDocument()
    expect(ch.queryByLabelText('Fórmula')).not.toBeInTheDocument()
  })

  it('acurácia diz exatamente como é (ou não) usada na pontuação', () => {
    renderAt('/calculos')
    const ch = within(document.getElementById('acuracia')!)
    expect(ch.getByText(/No MVP a acurácia não altera a nota de cada característica/)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Modo técnico' }))
    expect(within(document.getElementById('acuracia')!).getByLabelText('Penalidade por baixa acurácia')).toHaveTextContent('pontos')
  })

  it('datasets cita a mensagem de bases incompatíveis e os fatores de comparabilidade', () => {
    renderAt('/calculos')
    const ch = within(document.getElementById('datasets')!)
    expect(ch.getByText('⚠ Comparação genética não suportada entre essas bases.')).toBeInTheDocument()
    expect(ch.getByText(/raça, programa, release, base genética, característica, unidade e metodologia/)).toBeInTheDocument()
  })

  it('a lista de status do resumo reflete o estado real de cada capítulo', () => {
    renderAt('/calculos')
    const sum = within(screen.getByRole('region', { name: 'Resumo do que é calculado' }))
    const row = (title: string) => sum.getByRole('link', { name: title }).closest('li')!
    expect(row('Parentesco / endogamia')).toHaveTextContent('Planejado')
    expect(row('Progênie esperada')).toHaveTextContent('Não implementado')
    expect(row('Score de acasalamento')).toHaveTextContent('Implementado')
    expect(sum.getByText(/O MVP não possui validação zootécnica/)).toBeInTheDocument()
  })
})
