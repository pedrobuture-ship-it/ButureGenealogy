import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import App from '../App'
import { installFakeBackend, sampleAnimal, sampleFarm } from '../test/fakeBackend'

function open(hash: string) {
  window.location.hash = hash
  return render(<App />)
}

describe('Rebanho', () => {
  it('mostra estado vazio sem propriedade e permite cadastrar uma', async () => {
    const be = installFakeBackend()
    open('#/rebanho')
    expect(await screen.findByText('Nenhuma propriedade cadastrada')).toBeInTheDocument()

    fireEvent.click(screen.getAllByRole('button', { name: 'Nova propriedade' })[0])
    const dialog = await screen.findByRole('dialog', { name: 'Nova propriedade' })
    // validation error comes from the backend and is shown inline
    fireEvent.click(within(dialog).getByRole('button', { name: 'Salvar' }))
    expect(await within(dialog).findByText('Informe o nome da propriedade.')).toBeInTheDocument()

    fireEvent.change(within(dialog).getByLabelText('Nome'), { target: { value: 'Fazenda Nova' } })
    fireEvent.click(within(dialog).getByRole('button', { name: 'Salvar' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(be.farms).toHaveLength(1)
    expect(await screen.findByText('Nenhum animal cadastrado')).toBeInTheDocument()
  })

  it('lista animais, filtra por busca e abre os detalhes', async () => {
    installFakeBackend({ farms: [sampleFarm], animals: [sampleAnimal(), sampleAnimal({ id: 'a2', name: 'Estrela', registration: 'R-2' })] })
    open('#/rebanho')
    expect(await screen.findByText('Mimosa')).toBeInTheDocument()
    expect(screen.getByText('Estrela')).toBeInTheDocument()

    fireEvent.change(screen.getByLabelText('Buscar por nome ou registro'), { target: { value: 'estre' } })
    await waitFor(() => expect(screen.queryByText('Mimosa')).not.toBeInTheDocument())

    fireEvent.click(screen.getByText('Estrela'))
    expect(await screen.findByRole('heading', { level: 1, name: 'Estrela' })).toBeInTheDocument()
    expect(screen.getAllByRole('tab').map((t) => t.textContent)).toEqual(['Resumo', 'Fenótipo', 'Genética', 'Genealogia', 'Reprodução'])
  })

  it('cadastra um animal e navega para a ficha', async () => {
    const be = installFakeBackend({ farms: [sampleFarm] })
    open('#/rebanho')
    await screen.findByText('Nenhum animal cadastrado')
    fireEvent.click(screen.getByRole('button', { name: 'Novo animal' }))
    const dialog = await screen.findByRole('dialog', { name: 'Novo animal' })
    fireEvent.click(within(dialog).getByRole('button', { name: 'Salvar' }))
    expect(await within(dialog).findByText('Informe o nome ou o registro do animal.')).toBeInTheDocument()

    fireEvent.change(within(dialog).getByLabelText('Nome'), { target: { value: 'Bonita' } })
    fireEvent.click(within(dialog).getByRole('button', { name: 'Salvar' }))
    expect(await screen.findByRole('heading', { level: 1, name: 'Bonita' })).toBeInTheDocument()
    expect(be.animals).toHaveLength(1)
  })

  it('arquiva com confirmação e restaura', async () => {
    const be = installFakeBackend({ farms: [sampleFarm], animals: [sampleAnimal()] })
    open('#/rebanho/a1')
    await screen.findByRole('heading', { level: 1, name: 'Mimosa' })

    fireEvent.click(screen.getByRole('button', { name: 'Arquivar' }))
    const dialog = await screen.findByRole('dialog', { name: 'Arquivar animal' })
    // nothing happens until confirmed
    expect(be.animals[0].status).toBe('active')
    fireEvent.click(within(dialog).getByRole('button', { name: 'Arquivar' }))
    await waitFor(() => expect(be.animals[0].status).toBe('archived'))
    expect(await screen.findByRole('button', { name: 'Restaurar' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Restaurar' }))
    await waitFor(() => expect(be.animals[0].status).toBe('active'))
  })

  it('mostra erro compreensível para animal inexistente', async () => {
    installFakeBackend({ farms: [sampleFarm] })
    open('#/rebanho/nao-existe')
    expect(await screen.findByRole('alert')).toHaveTextContent('Animal não encontrado.')
  })
})
