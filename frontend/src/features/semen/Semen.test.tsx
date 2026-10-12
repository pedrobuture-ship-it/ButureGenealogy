import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import App from '../../App'
import { installFakeBackend, sampleAnimal, sampleFarm } from '../../test/fakeBackend'

const bull = (over = {}) => sampleAnimal({ id: 'b1', name: 'Touro Alfa', registration: 'T-1', sex: 'male', category: 'bull', ...over })
const lot = (over = {}) => ({ id: 'l1', bullId: 'b1', supplier: 'Central', batch: 'L1', quantity: 10, unitCost: 45, semenType: 'conventional', notes: '', ...over })

describe('Sêmen — banco de touros', () => {
  it('mostra as colunas combinadas e leva à ficha do touro', async () => {
    installFakeBackend({ farms: [sampleFarm], animals: [bull(), bull({ id: 'b2', name: 'Touro Sem Sêmen', registration: 'T-2' })], lots: [lot()] })
    window.location.hash = '#/semen'
    render(<App />)
    const row = (await screen.findByText('Touro Alfa')).closest('tr')!
    const heads = screen.getAllByRole('columnheader').map((h) => h.textContent?.replace(/[↑↓\s]+$/, '').trim())
    expect(heads).toEqual(['Touro', 'Raça', 'Registro', 'Fornecedor', 'Doses', 'Custo/dose', 'Status'])
    const cells = within(row).getAllByRole('cell')
    expect(cells[1]).toHaveTextContent('Nelore')
    expect(cells[2]).toHaveTextContent('T-1')
    expect(cells[3]).toHaveTextContent('Central')
    expect(cells[4]).toHaveTextContent('10')
    expect(cells[5]).toHaveTextContent(/R\$\s*45,00/)
    expect(cells[6]).toHaveTextContent('Em estoque')
    // bulls without semen are hidden by default and appear with "Todos os touros"
    expect(screen.queryByText('Touro Sem Sêmen')).not.toBeInTheDocument()
    fireEvent.change(screen.getByLabelText('Estoque'), { target: { value: 'all' } })
    const other = (await screen.findByText('Touro Sem Sêmen')).closest('tr')!
    expect(within(other).getByText('Sem sêmen')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Touro Alfa'))
    expect(await screen.findByRole('heading', { level: 1, name: 'Touro Alfa' })).toBeInTheDocument()
    expect(screen.getAllByRole('tab').map((t) => t.textContent)).toEqual(['Estoque', 'Genética', 'Genealogia', 'Movimentações'])
  })

  it('"Adicionar sêmen" busca um touro existente e abre o novo lote', async () => {
    installFakeBackend({ farms: [sampleFarm], animals: [bull()], lots: [] })
    window.location.hash = '#/semen'
    render(<App />)
    await screen.findByText('Nenhum touro encontrado')
    fireEvent.click(screen.getByRole('button', { name: 'Adicionar sêmen' }))
    const dlg = await screen.findByRole('dialog', { name: 'Adicionar sêmen' })
    fireEvent.change(within(dlg).getByLabelText('Buscar touro cadastrado'), { target: { value: 'alfa' } })
    fireEvent.click(await within(dlg).findByRole('button', { name: /Touro Alfa/ }))
    expect(await screen.findByRole('dialog', { name: 'Novo lote de sêmen' })).toBeInTheDocument()
  })
})

describe('Sêmen — ficha do touro', () => {
  const setup = () => {
    const be = installFakeBackend({ farms: [sampleFarm], animals: [bull()], lots: [lot()] })
    window.location.hash = '#/semen/b1'
    render(<App />)
    return be
  }

  it('retirada acima do estoque é recusada com mensagem clara; abaixo, atualiza o saldo e o histórico', async () => {
    const be = setup()
    expect(await screen.findByLabelText('Doses em estoque')).toHaveTextContent('10')
    fireEvent.click(screen.getByRole('button', { name: 'Retirar' }))
    const dlg = await screen.findByRole('dialog', { name: 'Retirar doses' })
    expect(within(dlg).getByText('Disponível neste lote: 10')).toBeInTheDocument()

    fireEvent.change(within(dlg).getByLabelText('Quantidade de doses'), { target: { value: '11' } })
    fireEvent.click(within(dlg).getByRole('button', { name: 'Confirmar' }))
    expect(await within(dlg).findByText('Estoque insuficiente: há 10 dose(s) neste lote.')).toBeInTheDocument()
    expect(be.lots[0].quantity).toBe(10)

    fireEvent.change(within(dlg).getByLabelText('Quantidade de doses'), { target: { value: '3' } })
    fireEvent.click(within(dlg).getByRole('button', { name: 'Confirmar' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(await screen.findByText('7', { selector: '.big-number' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('tab', { name: 'Movimentações' }))
    const row = (await screen.findByText('Retirada')).closest('tr')!
    const cells = within(row).getAllByRole('cell')
    expect(cells[3]).toHaveTextContent('-3')
    expect(cells[4]).toHaveTextContent('7')
    expect(cells[5]).toHaveTextContent('Inseminação')
  })

  it('valida a quantidade antes de enviar e exige observação no ajuste', async () => {
    const be = setup()
    fireEvent.click(await screen.findByRole('button', { name: 'Retirar' }))
    const dlg = await screen.findByRole('dialog', { name: 'Retirar doses' })
    fireEvent.click(within(dlg).getByRole('button', { name: 'Confirmar' }))
    expect(await within(dlg).findByText('Informe um número inteiro maior que zero.')).toBeInTheDocument()
    expect(be.calls.some((c) => c.method === 'WithdrawSemen')).toBe(false)
    fireEvent.click(within(dlg).getByRole('button', { name: 'Cancelar' }))

    fireEvent.click(screen.getByRole('button', { name: 'Ajustar' }))
    const adj = await screen.findByRole('dialog', { name: 'Ajustar contagem' })
    expect(within(adj).queryByLabelText('Motivo')).not.toBeInTheDocument()
    expect(within(adj).getByLabelText('Observação')).toBeInTheDocument()
  })

  it('cadastra um novo lote com fornecedor, lote, tipo, preço e doses', async () => {
    const be = setup()
    fireEvent.click(await screen.findByRole('button', { name: 'Novo lote' }))
    const dlg = await screen.findByRole('dialog', { name: 'Novo lote de sêmen' })
    fireEvent.change(within(dlg).getByLabelText(/Fornecedor/), { target: { value: 'Outra Central' } })
    fireEvent.change(within(dlg).getByLabelText(/^Lote/), { target: { value: 'Z9' } })
    fireEvent.change(within(dlg).getByLabelText('Tipo de sêmen'), { target: { value: 'sexed' } })
    fireEvent.change(within(dlg).getByLabelText(/Preço por dose/), { target: { value: '89,90' } })
    fireEvent.change(within(dlg).getByLabelText(/Doses iniciais/), { target: { value: '25' } })
    fireEvent.click(within(dlg).getByRole('button', { name: 'Salvar' }))
    await waitFor(() => expect(be.lots).toHaveLength(2))
    expect(be.lots[1]).toMatchObject({ supplier: 'Outra Central', batch: 'Z9', semenType: 'sexed', quantity: 25, unitCost: 89.9 })
  })

  it('a ficha do touro mostra dataset, versão, acurácia e percentil na aba Genética', async () => {
    const prov = { kindCode: 'DEP', kindLabel: 'DEP', genetic: true, traitCode: 'WW', traitName: 'Peso à desmama', datasetId: 'd1', datasetName: 'Nelore Example', datasetVersion: '2026.1',
      program: 'Prog', release: 'r', geneticBase: '', sourceReference: '', accuracy: 0.88, source: '', date: '', isDemo: false }
    installFakeBackend({ farms: [sampleFarm], animals: [bull()], lots: [lot()], traitGroups: [{ category: 'growth', items: [
      { id: 'e1', source: 'evaluation', traitCode: 'WW', traitName: 'Peso à desmama', defined: true, category: 'growth', valueKind: 'DEP', unit: 'kg', value: 6.1, accuracy: 0.88, percentile: 82, notes: '', optimizationDirection: 'INFORMATION_ONLY', provenance: prov }] }] })
    window.location.hash = '#/semen/b1'
    render(<App />)
    fireEvent.click(await screen.findByRole('tab', { name: 'Genética' }))
    const section = await screen.findByRole('region', { name: 'Crescimento' })
    const cells = within(section).getAllByRole('row')[1].querySelectorAll('td')
    expect(cells[4]).toHaveTextContent('88%')
    expect(cells[5]).toHaveTextContent('82')
    expect(cells[6]).toHaveTextContent('Nelore Example · 2026.1')
  })
})
