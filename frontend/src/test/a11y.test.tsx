import axe from 'axe-core'
import { fireEvent, render, screen } from '@testing-library/react'
import App from '../App'
import { installFakeBackend, sampleAnimal, sampleFarm } from './fakeBackend'
import { profileOf, runResponse, specSnapshot } from './engineFixtures'

async function violations(container: HTMLElement) {
  const r = await axe.run(container, { rules: { 'color-contrast': { enabled: false } } }) // jsdom has no layout/colors
  return r.violations.map((v) => `${v.id}: ${v.help} (${v.nodes.map((n) => n.target.join(' ')).join(' | ')})`)
}

const lot = { id: 'l1', bullId: 'b', supplier: 'Central', batch: 'L1', quantity: 10, unitCost: 45, semenType: 'conventional', notes: '' }
const animals = [sampleAnimal(), sampleAnimal({ id: 'b', name: 'Touro A', sex: 'male', category: 'bull' })]

const pages: [string, string, () => Promise<void> | void][] = [
  ['Início', '#/', async () => { await screen.findByText('Saudável') }],
  ['Rebanho', '#/rebanho', async () => { await screen.findByText('Mimosa') }],
  ['Ficha do animal', '#/rebanho/a1', async () => { await screen.findByRole('heading', { level: 1, name: 'Mimosa' }) }],
  ['Genealogia', '#/genealogia/a1', async () => { await screen.findByRole('tree') }],
  ['Sêmen', '#/semen', async () => { await screen.findByText('Touro A') }],
  ['Ficha do touro', '#/semen/b', async () => { await screen.findByLabelText('Doses em estoque') }],
  ['Bibliotecas', '#/bibliotecas', async () => { await screen.findByText('Nenhuma biblioteca instalada') }],
  ['Acasalamentos', '#/acasalamentos', async () => { await screen.findByRole('button', { name: 'Calcular ranking' }) }],
  ['Cálculos', '#/calculos', async () => { await screen.findByRole('navigation', { name: 'Capítulos' }) }],
  ['Configurações', '#/configuracoes', async () => { await screen.findByRole('button', { name: 'Criar backup' }) }],
]

describe('acessibilidade (axe)', () => {
  for (const [name, hash, ready] of pages) {
    it(`${name} não tem violações detectáveis`, async () => {
      installFakeBackend({ farms: [sampleFarm], animals, lots: [lot], profiles: [profileOf()], runResponse: runResponse(), calculations: [{ id: 'm9', snapshot: specSnapshot() }] })
      window.location.hash = hash
      const { container } = render(<App />)
      await ready()
      expect(await violations(container)).toEqual([])
    })
  }

  it('janelas de diálogo (modal e drawer) são acessíveis', async () => {
    installFakeBackend({ farms: [sampleFarm], animals, lots: [lot], profiles: [profileOf()], runResponse: runResponse() })
    window.location.hash = '#/rebanho'
    const { container } = render(<App />)
    await screen.findByText('Mimosa')
    fireEvent.click(screen.getByRole('button', { name: 'Novo animal' }))
    await screen.findByRole('dialog', { name: 'Novo animal' })
    expect(await violations(container)).toEqual([])
  })
})
