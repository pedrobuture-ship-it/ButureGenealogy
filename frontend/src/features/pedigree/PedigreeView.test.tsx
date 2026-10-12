import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import App from '../../App'
import { installFakeBackend, sampleAnimal, sampleFarm } from '../../test/fakeBackend'

const herd = () => [
  sampleAnimal({ id: 'calf', name: 'Bezerra', registration: '', fatherId: 'sire', motherId: '' }),
  sampleAnimal({ id: 'sire', name: 'Touro Pai', sex: 'male', category: 'bull' }),
]

describe('Genealogia', () => {
  it('mostra a árvore, destaca ancestral desconhecido e navega ao clicar em um ancestral', async () => {
    const be = installFakeBackend({ farms: [sampleFarm], animals: herd() })
    window.location.hash = '#/genealogia/calf'
    render(<App />)

    const tree = await screen.findByRole('tree', { name: 'Árvore genealógica' })
    expect(within(tree).getByText('Bezerra')).toBeInTheDocument()
    expect(within(tree).getByText('Touro Pai')).toBeInTheDocument()
    expect(within(tree).getByRole('treeitem', { name: 'Cadastrar mãe de Bezerra' })).toBeInTheDocument()
    expect(screen.getByText('não calculado nesta versão')).toBeInTheDocument()

    // generation selector re-queries the backend
    fireEvent.click(screen.getByRole('button', { name: '4 gerações' }))
    await waitFor(() => expect(be.calls.some((c) => c.method === 'GetPedigree' && c.args[1] === 4)).toBe(true))

    fireEvent.click(within(tree).getByText('Touro Pai'))
    expect(await screen.findByRole('heading', { level: 1, name: 'Touro Pai' })).toBeInTheDocument()
  })

  it('cria um registro mínimo para o ancestral desconhecido', async () => {
    const be = installFakeBackend({ farms: [sampleFarm], animals: herd() })
    window.location.hash = '#/genealogia/calf'
    render(<App />)

    fireEvent.click(await screen.findByRole('treeitem', { name: 'Cadastrar mãe de Bezerra' }))
    const dialog = await screen.findByRole('dialog', { name: 'Cadastrar mãe' })
    fireEvent.click(within(dialog).getByRole('button', { name: 'Salvar' }))
    expect(await within(dialog).findByText('Informe o nome ou o registro do animal.')).toBeInTheDocument()

    fireEvent.change(within(dialog).getByLabelText('Nome'), { target: { value: 'Vaca Mãe' } })
    fireEvent.click(within(dialog).getByRole('button', { name: 'Salvar' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(be.animals.find((a) => a.id === 'calf')!.motherId).not.toBe('')
    expect(await screen.findByText('Vaca Mãe')).toBeInTheDocument()
  })

  it('a aba Genealogia da ficha exibe a árvore', async () => {
    installFakeBackend({ farms: [sampleFarm], animals: herd() })
    window.location.hash = '#/rebanho/calf'
    render(<App />)
    await screen.findByRole('heading', { level: 1, name: 'Bezerra' })
    fireEvent.click(screen.getByRole('tab', { name: 'Genealogia' }))
    expect(await screen.findByRole('tree')).toBeInTheDocument()
  })
})
