import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import App from '../App'
import { installFakeBackend, sampleAnimal, sampleFarm } from '../test/fakeBackend'
import { demoNotice, profileOf, runResponse, specSnapshot } from '../test/engineFixtures'

const lot = { id: 'l1', bullId: 'b', supplier: 'Central', batch: 'L', quantity: 10, unitCost: 45, semenType: 'conventional', notes: '' }
const bullAnimals = [sampleAnimal({ id: 'b', name: 'Touro A', sex: 'male', category: 'bull' }), sampleAnimal({ id: 'b2', name: 'Touro B', sex: 'male', category: 'bull' })]
const profiles = () => [
  profileOf(),
  profileOf({ id: 'builtin:demo-beef', name: 'Demonstração Corte', isDemo: true, demoNotice, traits: [
    { traitCode: 'WW', evaluationType: 'DEP', weight: 40, direction: 'MAXIMIZE', idealMin: null, idealMax: null, targetTolerance: null, required: false }] }),
]

function setup(extra: Record<string, unknown> = {}) {
  const be = installFakeBackend({
    farms: [sampleFarm], animals: [sampleAnimal({ id: 'f', name: 'Matriz 0472' }), ...bullAnimals],
    lots: [lot, { ...lot, id: 'l2', bullId: 'b2' }], profiles: profiles(), runResponse: runResponse(),
    calculations: [{ id: 'm9', snapshot: specSnapshot(), score: 65.8 }], ...extra,
  })
  window.location.hash = '#/acasalamentos'
  render(<App />)
  return be
}

async function chooseFemaleAndBulls() {
  fireEvent.change(await screen.findByLabelText('Buscar matriz'), { target: { value: 'Matriz' } })
  fireEvent.click(await screen.findByRole('button', { name: /Matriz 0472/ }))
  fireEvent.click(await screen.findByLabelText('Selecionar Touro A'))
  fireEvent.click(screen.getByLabelText('Selecionar Touro B'))
}

describe('Acasalamentos', () => {
  it('só permite calcular com matriz, touros e perfil escolhidos', async () => {
    setup()
    const btn = await screen.findByRole('button', { name: 'Calcular ranking' })
    expect(btn).toBeDisabled()
    await chooseFemaleAndBulls()
    await waitFor(() => expect(btn).toBeEnabled())
    expect(screen.getByText('2 touro(s) selecionado(s)')).toBeInTheDocument()
  })

  it('envia exatamente a matriz, os touros e os pesos editados pelo usuário, e mostra ranking, top e avisos', async () => {
    const be = setup()
    await chooseFemaleAndBulls()
    fireEvent.change(screen.getByLabelText('Perfil de objetivo'), { target: { value: 'builtin:demo-beef' } })
    // DEMO profiles carry the mandatory notice
    expect(await screen.findAllByText(demoNotice)).not.toHaveLength(0)
    // every weight is editable
    fireEvent.change(await screen.findByLabelText('Peso de WW'), { target: { value: '55' } })
    fireEvent.click(screen.getByRole('button', { name: 'Calcular ranking' }))

    await screen.findByRole('region', { name: 'Ranking' })
    const sent = be.calls.find((c) => c.method === 'RunMating')!.args[0] as { femaleId: string; bullIds: string[]; save: boolean; profile: { id: string; traits: { weight: number }[] } }
    expect(sent.femaleId).toBe('f')
    expect([...sent.bullIds].sort()).toEqual(['b', 'b2'])
    expect(sent.save).toBe(true)
    expect(sent.profile.traits[0].weight).toBe(55)

    // top cards, ranking order and final scores come from the response
    const top = within(screen.getByRole('region', { name: 'Top candidatos' }))
    expect(top.getByText('Touro A')).toBeInTheDocument()
    expect(top.getAllByText('65,80')).toHaveLength(1)
    const rows = within(screen.getByRole('region', { name: 'Ranking' })).getAllByRole('row').slice(1)
    expect(rows).toHaveLength(2)
    expect(rows[0]).toHaveTextContent('Touro A')
    expect(rows[0]).toHaveTextContent('65,80')
    expect(rows[1]).toHaveTextContent('Touro B')
    expect(rows[1]).toHaveTextContent('60,50')

    // not calculable pairs explain themselves instead of showing a number
    const nc = within(screen.getByRole('region', { name: 'Não calculáveis' }))
    expect(nc.getByText('Touro Gama')).toBeInTheDocument()
    expect(nc.getAllByText(/Comparação genética não suportada entre essas bases/).length).toBeGreaterThan(0)
    // excluded by filters, with the reason
    const ex = within(screen.getByRole('region', { name: 'Excluídos pelos filtros' }))
    expect(ex.getByText('Touro Sem Estoque')).toBeInTheDocument()
    expect(ex.getByText(/Estoque insuficiente/)).toBeInTheDocument()
  })

  it('"Ver cálculo" abre o drawer com o snapshot do par escolhido', async () => {
    setup()
    await chooseFemaleAndBulls()
    fireEvent.click(screen.getByRole('button', { name: 'Calcular ranking' }))
    const ranking = within(await screen.findByRole('region', { name: 'Ranking' }))
    fireEvent.click(ranking.getAllByRole('button', { name: 'Ver cálculo' })[0])
    const dlg = await screen.findByRole('dialog', { name: 'Ver cálculo' })
    expect(within(dlg).getByLabelText('Score final')).toHaveTextContent('65,80 / 100')
    expect(within(dlg).getByText('Touro A', { exact: false })).toBeInTheDocument()
    fireEvent.keyDown(document, { key: 'Escape' })
    await waitFor(() => expect(screen.queryByRole('dialog', { name: 'Ver cálculo' })).not.toBeInTheDocument())
  })

  it('o histórico abre cálculos antigos exatamente como foram salvos', async () => {
    const be = setup()
    fireEvent.change(await screen.findByLabelText('Buscar matriz'), { target: { value: 'Matriz' } })
    fireEvent.click(await screen.findByRole('button', { name: /Matriz 0472/ }))
    const hist = await screen.findByRole('region', { name: 'Cálculos anteriores' })
    fireEvent.click(within(hist).getByRole('button', { name: 'Ver cálculo' }))
    const dlg = await screen.findByRole('dialog', { name: 'Ver cálculo' })
    expect(await within(dlg).findByLabelText('Score final')).toHaveTextContent('65,80')
    expect(be.calls.some((c) => c.method === 'GetCalculation' && c.args[0] === 'm9')).toBe(true)
  })

  it('erros do motor são mostrados e nada é apresentado como resultado', async () => {
    const resp = runResponse()
    resp.output.errors = ['O perfil não tem nenhuma característica com peso maior que zero.']
    resp.output.results = []
    resp.saved = false
    setup({ runResponse: resp })
    await chooseFemaleAndBulls()
    fireEvent.click(screen.getByRole('button', { name: 'Calcular ranking' }))
    expect(await screen.findByText('O perfil não tem nenhuma característica com peso maior que zero.')).toBeInTheDocument()
    expect(screen.queryByRole('region', { name: 'Ranking' })).not.toBeInTheDocument()
  })

  it('uma característica sem direção definida avisa o usuário', async () => {
    setup({ profiles: [profileOf({ traits: [{ traitCode: 'WW', evaluationType: 'DEP', weight: 10, direction: '', idealMin: null, idealMax: null, targetTolerance: null, required: false }] })] })
    expect(await screen.findByText(/Sem direção definida/)).toBeInTheDocument()
  })
})
