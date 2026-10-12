import { render, screen, within } from '@testing-library/react'
import type { Provenance } from '../types/domain'
import { DataProvenance } from './DataProvenance'

const dep: Provenance = {
  kindCode: 'DEP', kindLabel: 'DEP', genetic: true, traitCode: 'WW', traitName: 'Peso à desmama',
  datasetId: 'd1', datasetName: 'Dataset X', datasetVersion: 'Y', program: 'Z', release: '2026.1',
  geneticBase: 'base 2020', sourceReference: '', accuracy: 0.81, source: '', date: '', isDemo: false,
}

describe('DataProvenance', () => {
  it('mostra característica, tipo, dataset, versão, programa e acurácia', () => {
    render(<DataProvenance provenance={dep} variant="block" />)
    const rows = screen.getByText('Característica').closest('dl')!
    for (const [k, v] of [['Característica', 'Peso à desmama'], ['Tipo', 'DEP'], ['Dataset', 'Dataset X'], ['Versão', 'Y'], ['Programa', 'Z'], ['Acurácia', '81%'], ['Base genética', 'base 2020']]) {
      expect(within(rows).getByText(k)).toBeInTheDocument()
      expect(within(rows).getByText(v)).toBeInTheDocument()
    }
  })

  it('fenótipo não exibe campos genéticos e dado DEMO é sinalizado', () => {
    const pheno: Provenance = { ...dep, kindCode: 'PHENOTYPE', kindLabel: 'Fenótipo', genetic: false, datasetName: '', program: '', accuracy: null, source: 'balança', isDemo: true }
    render(<DataProvenance provenance={pheno} variant="block" />)
    expect(screen.queryByText('Programa')).not.toBeInTheDocument()
    expect(screen.queryByText('Acurácia')).not.toBeInTheDocument()
    expect(screen.getByText('balança')).toBeInTheDocument()
    expect(screen.getByText(/dado fictício/)).toBeInTheDocument()
  })

  it('na variante inline fica recolhida atrás de "Origem"', () => {
    render(<DataProvenance provenance={dep} />)
    expect(screen.getByText('Origem')).toBeInTheDocument()
  })
})
