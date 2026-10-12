import { fireEvent, render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { notCalculableSnapshot, profileOf, specSnapshot, demoNotice } from '../../test/engineFixtures'
import { CalculationView } from './CalculationView'

const renderView = (snapshot = specSnapshot()) => render(<MemoryRouter><CalculationView snapshot={snapshot} /></MemoryRouter>)

describe('Ver cálculo', () => {
  it('mostra o score final e os componentes com score normalizado, peso e contribuição do snapshot', () => {
    renderView()
    expect(screen.getByLabelText('Score final')).toHaveTextContent('65,80 / 100')
    const comps = within(screen.getByRole('region', { name: 'Componentes' }))
    const fert = comps.getByText('Fertilidade (DEP)').closest('article')!
    expect(within(fert).getByText('Score normalizado').nextElementSibling).toHaveTextContent('82,00')
    expect(within(fert).getByText('Peso').nextElementSibling).toHaveTextContent('30%')
    expect(within(fert).getByText('Contribuição').nextElementSibling).toHaveTextContent('24,60')
    const carc = comps.getByText('Carcaça (DEP)').closest('article')!
    expect(within(carc).getByText('Contribuição').nextElementSibling).toHaveTextContent('22,75')
    const mat = comps.getByText('Maternal (DEP)').closest('article')!
    expect(within(mat).getByText('Peso').nextElementSibling).toHaveTextContent('20%')
    expect(within(mat).getByText('Contribuição').nextElementSibling).toHaveTextContent('15,20')
  })

  it('decompõe o resultado como 64,00 + 4,10 − 2,30 = 65,80 sem recalcular', () => {
    // the fixture is deliberately NOT internally consistent with its components
    // (24,60+22,75+15,20 ≠ 64,00): the view must print the snapshot's numbers, not its own sum.
    renderView()
    const result = within(screen.getByRole('region', { name: 'Resultado' }))
    expect(result.getByText('Soma dos componentes').nextElementSibling).toHaveTextContent('64,00')
    expect(result.getByText('Complementaridade').nextElementSibling).toHaveTextContent('+4,10')
    expect(result.getByText('Penalidades').nextElementSibling).toHaveTextContent('−2,30')
    expect(result.getByText('Score final').nextElementSibling).toHaveTextContent('= 65,80')
    expect(within(screen.getByRole('region', { name: 'Complementaridade' })).getByText('+4,10', { selector: '.calc-num' })).toBeInTheDocument()
    expect(within(screen.getByRole('region', { name: 'Penalidades' })).getByText('−2,30', { selector: '.calc-num' })).toBeInTheDocument()
  })

  it('"Como chegamos nisso?" abre valor da matriz e do touro, tipo, dataset, versão, unidade, transformação, fórmula, peso e resultado', () => {
    renderView()
    const fert = within(screen.getByRole('region', { name: 'Componentes' })).getByText('Fertilidade (DEP)').closest('article')!
    expect(within(fert).queryByText('Valor da matriz')).not.toBeInTheDocument()
    fireEvent.click(within(fert).getByRole('button', { name: 'Como chegamos nisso?' }))
    const how = within(fert.querySelector('.how') as HTMLElement)
    for (const label of ['Valor da matriz', 'Valor do touro', 'Tipo da avaliação', 'Dataset', 'Versão', 'Unidade', 'Transformação', 'Fórmula / método', 'Peso', 'Resultado']) {
      expect(how.getByText(label, { selector: 'dt' })).toBeInTheDocument()
    }
    expect(how.getByText('Valor da matriz').nextElementSibling).toHaveTextContent('2,2 kg')
    expect(how.getByText('Valor do touro').nextElementSibling).toHaveTextContent('6,1 kg')
    expect(how.getByText('Tipo da avaliação').nextElementSibling).toHaveTextContent('DEP')
    expect(how.getByText('Dataset', { selector: 'dt' }).nextElementSibling).toHaveTextContent('Nelore Example')
    expect(how.getByText('Versão', { selector: 'dt' }).nextElementSibling).toHaveTextContent('2026.1')
    expect(how.getByText('Resultado', { selector: 'dt' }).nextElementSibling).toHaveTextContent('82,00 × 30% = 24,60')
    expect(how.getByRole('link', { name: 'normalização' })).toHaveAttribute('href', '/calculos/normalizacao')
  })

  it('mostra "Dados usados neste cálculo": dataset, versão, motor, configuração e data', () => {
    renderView()
    const data = within(screen.getByRole('region', { name: 'Dados usados neste cálculo' }))
    expect(data.getByText('Dataset', { selector: 'dt' }).nextElementSibling).toHaveTextContent('Nelore Example')
    expect(data.getByText('Versão', { selector: 'dt' }).nextElementSibling).toHaveTextContent('2026.1')
    expect(data.getByText('Motor').nextElementSibling).toHaveTextContent('0.1.0')
    expect(data.getByText('Configuração').nextElementSibling).toHaveTextContent('Personalizada')
    expect(data.getByText('Executado em').nextElementSibling).toHaveTextContent('11/10/2026')
  })

  it('liga a explicações da página Cálculos', () => {
    renderView()
    expect(screen.getAllByRole('link').map((a) => a.getAttribute('href'))).toEqual(expect.arrayContaining(['/calculos/pesos', '/calculos/complementaridade', '/calculos/parentesco']))
  })

  it('endogamia aparece como "Planejado — ainda não utilizado no score do MVP", sem valor', () => {
    renderView()
    const pen = within(screen.getByRole('region', { name: 'Penalidades' }))
    expect(pen.getByText(/Planejado — ainda não utilizado no score do MVP/)).toBeInTheDocument()
    expect(pen.getByText('não aplicada')).toBeInTheDocument()
  })

  it('sinaliza dados DEMO claramente', () => {
    const s = specSnapshot({ isDemo: true, profile: profileOf({ isDemo: true, demoNotice, name: 'Demonstração Corte' }) })
    s.datasets = s.datasets.map((d) => ({ ...d, isDemo: true }))
    renderView(s)
    expect(screen.getByRole('note')).toHaveTextContent('DEMO / SYNTHETIC')
    expect(screen.getByRole('note')).toHaveTextContent(demoNotice)
    expect(screen.getAllByText('DEMO').length).toBeGreaterThan(0)
  })

  it('par não calculável mostra os motivos e o aviso de bases incompatíveis, sem score', () => {
    renderView(notCalculableSnapshot())
    expect(screen.getByLabelText('Score final')).toHaveTextContent('Não calculável')
    expect(screen.getByText(/Apenas 80,00% do peso/)).toBeInTheDocument()
    expect(screen.getByText('⚠ Comparação genética não suportada entre essas bases.')).toBeInTheDocument()
    expect(screen.queryByRole('region', { name: 'Resultado' })).not.toBeInTheDocument()
  })

  it('score limitado a 0–100 informa o valor sem limite', () => {
    const s = specSnapshot()
    s.result.final = { ...s.result.final!, unclamped: 109.45, final: 100, clamped: true }
    renderView(s)
    expect(screen.getByText(/limitado ao intervalo de 0 a 100 \(valor sem limite: 109,45\)/)).toBeInTheDocument()
  })
})
