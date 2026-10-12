import { fireEvent, render, screen, within } from '@testing-library/react'
import App from '../../App'
import { installFakeBackend, sampleAnimal, sampleFarm } from '../../test/fakeBackend'

const prov = (over = {}) => ({ kindCode: 'DEP', kindLabel: 'DEP', genetic: true, traitCode: 'WW', traitName: 'Peso à desmama', datasetId: '', datasetName: '',
  datasetVersion: '', program: 'Prog', release: 'r1', geneticBase: '', sourceReference: '', accuracy: 0.8, source: '', date: '', isDemo: false, ...over })

const groups = [
  { category: 'growth', items: [
    { id: 'e1', source: 'evaluation', traitCode: 'WW', traitName: 'Peso à desmama', defined: true, category: 'growth', valueKind: 'DEP', unit: 'kg', value: 5, accuracy: 0.8, percentile: 70, notes: '', optimizationDirection: 'INFORMATION_ONLY', provenance: prov() },
    { id: 'e2', source: 'evaluation', traitCode: 'WW', traitName: 'Peso à desmama', defined: true, category: 'growth', valueKind: 'EBV', unit: 'kg', value: 11, accuracy: null, percentile: null, notes: '', optimizationDirection: 'INFORMATION_ONLY', provenance: prov({ kindCode: 'EBV', kindLabel: 'EBV', accuracy: null }) },
    { id: 'm1', source: 'measurement', traitCode: 'BODY_WEIGHT', traitName: 'Peso corporal', defined: true, category: 'growth', valueKind: 'PHENOTYPE', unit: 'kg', value: 480, accuracy: null, percentile: null, notes: '', optimizationDirection: 'INFORMATION_ONLY', provenance: prov({ kindCode: 'PHENOTYPE', kindLabel: 'Fenótipo', traitName: 'Peso corporal', genetic: false, source: 'balança', program: '' }) },
  ] },
]

describe('abas Fenótipo e Genética', () => {
  it('Genética agrupa por categoria e mantém DEP e EBV como linhas separadas', async () => {
    installFakeBackend({ farms: [sampleFarm], animals: [sampleAnimal()], traitGroups: groups })
    window.location.hash = '#/rebanho/a1'
    render(<App />)
    await screen.findByRole('heading', { level: 1, name: 'Mimosa' })
    fireEvent.click(screen.getByRole('tab', { name: 'Genética' }))
    const section = await screen.findByRole('region', { name: 'Crescimento' })
    const rows = within(section).getAllByRole('row').slice(1)
    expect(rows).toHaveLength(2) // phenotype is not shown here
    expect(within(rows[0]).getAllByRole('cell')[1]).toHaveTextContent('DEP')
    expect(within(rows[1]).getAllByRole('cell')[1]).toHaveTextContent('EBV')
    const cells = within(rows[0]).getAllByRole('cell')
    expect(cells[4]).toHaveTextContent('80%') // acurácia
    expect(cells[5]).toHaveTextContent('70') // percentil
  })

  it('Fenótipo mostra só medições e oferece registrar nova', async () => {
    installFakeBackend({ farms: [sampleFarm], animals: [sampleAnimal()], traitGroups: groups })
    window.location.hash = '#/rebanho/a1'
    render(<App />)
    await screen.findByRole('heading', { level: 1, name: 'Mimosa' })
    fireEvent.click(screen.getByRole('tab', { name: 'Fenótipo' }))
    const section = await screen.findByRole('region', { name: 'Crescimento' })
    const rows = within(section).getAllByRole('row').slice(1)
    expect(rows).toHaveLength(1)
    expect(rows[0]).toHaveTextContent('Peso corporal')
    expect(section).not.toHaveTextContent('Peso à desmama')
    fireEvent.click(screen.getByRole('button', { name: 'Registrar medição' }))
    const form = await screen.findByRole('form', { name: 'Registrar medição' })
    expect(await within(form).findByRole('option', { name: 'Peso corporal (kg)' })).toBeInTheDocument()
  })
})
