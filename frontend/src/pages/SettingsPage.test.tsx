import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import App from '../App'
import { installFakeBackend, sampleFarm } from '../test/fakeBackend'

const preview = (over = {}) => ({
  path: '/home/u/antigo.genbackup', upgradedFrom: 6, schemaVersion: 6,
  manifest: { formatVersion: 1, createdAt: '2026-10-01T12:00:00Z', appVersion: '0.1.0', engineVersion: '0.1.0', schemaVersion: 6, farms: ['Fazenda São João'], counts: { animals: 120, farms: 1 }, files: [],
    libraries: [{ id: 'l1', packageId: 'x', name: 'Minha Lib', version: '1.0', fileName: 'a.genlib', sha256: '', licenseStatus: 'unknown', isDemo: false, included: true }, { id: 'l2', packageId: 'y', name: 'Pública', version: '2', fileName: 'b.genlib', sha256: '', licenseStatus: 'redistributable', isDemo: false, included: false }] },
  ...over,
})

const open = () => { window.location.hash = '#/configuracoes'; render(<App />) }

describe('Configurações — backup', () => {
  it('cria backup com o nome sugerido e mostra onde foi salvo', async () => {
    const be = installFakeBackend({ farms: [sampleFarm] })
    open()
    fireEvent.click(await screen.findByRole('button', { name: 'Criar backup' }))
    expect(await screen.findByText('/home/u/2026-10-11-fazenda.genbackup')).toBeInTheDocument()
    expect(be.calls.find((c) => c.method === 'PickBackupSavePath')!.args[0]).toBe('2026-10-11-fazenda.genbackup')
    expect(be.calls.find((c) => c.method === 'CreateBackup')!.args[0]).toBe('/home/u/2026-10-11-fazenda.genbackup')
    expect(screen.getByText(/12 animais/)).toBeInTheDocument()
  })

  it('restaurar valida primeiro, mostra o resumo e só substitui com confirmação explícita', async () => {
    const be = installFakeBackend({ farms: [sampleFarm], inspectBackup: preview() })
    open()
    fireEvent.click(await screen.findByRole('button', { name: 'Restaurar backup…' }))
    const dlg = await screen.findByRole('dialog', { name: 'Restaurar backup' })
    expect(be.calls.some((c) => c.method === 'InspectBackup')).toBe(true)
    expect(be.calls.some((c) => c.method === 'RestoreBackup')).toBe(false) // nothing happens before confirming
    expect(within(dlg).getByText('Fazenda São João')).toBeInTheDocument()
    expect(within(dlg).getByText('120')).toBeInTheDocument()
    expect(within(dlg).getByText(/Pública 2 \(somente metadados\)/)).toBeInTheDocument()
    expect(within(dlg).getByText(/backup de segurança da base atual será criado automaticamente/)).toBeInTheDocument()

    // without the checkbox the destructive action is refused
    fireEvent.click(within(dlg).getByRole('button', { name: 'Substituir os dados atuais' }))
    expect(await within(dlg).findByText('Marque a confirmação para continuar.')).toBeInTheDocument()
    expect(be.calls.some((c) => c.method === 'RestoreBackup')).toBe(false)

    fireEvent.click(within(dlg).getByLabelText(/Entendo que os dados atuais serão substituídos/))
    fireEvent.click(within(dlg).getByRole('button', { name: 'Substituir os dados atuais' }))
    await waitFor(() => expect(be.calls.some((c) => c.method === 'RestoreBackup' && c.args[0] === '/home/u/antigo.genbackup')).toBe(true))
    expect(await screen.findByText(/pre-restore-20261011-120000\.genbackup/)).toBeInTheDocument()
  })

  it('avisa quando o backup é de uma versão anterior e será atualizado', async () => {
    installFakeBackend({ farms: [sampleFarm], inspectBackup: preview({ upgradedFrom: 3, schemaVersion: 6 }) })
    open()
    fireEvent.click(await screen.findByRole('button', { name: 'Restaurar backup…' }))
    expect(await screen.findByText(/versão anterior \(esquema 3\) e será atualizado para a versão atual \(6\)/)).toBeInTheDocument()
  })

  it('mostra o motivo quando o backup é recusado e não abre a confirmação', async () => {
    const be = installFakeBackend({ farms: [sampleFarm] })
    be.calls.length = 0
    window.go!.main!.App!.InspectBackup = async () => Promise.reject(JSON.stringify({ code: 'validation', message: 'backup inválido: o arquivo "farm.db" está corrompido ou foi alterado (SHA-256 não confere)' }))
    open()
    fireEvent.click(await screen.findByRole('button', { name: 'Restaurar backup…' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('SHA-256 não confere')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})

describe('Configurações — integridade, demo, atalhos', () => {
  it('relata erros e avisos de forma compreensível, e o caso sem problemas', async () => {
    installFakeBackend({ farms: [sampleFarm], integrity: { ok: true, checkedAt: '2026-10-11T12:00:00Z', errors: 0, warnings: 0, issues: [], checked: {} } })
    open()
    fireEvent.click(await screen.findByRole('button', { name: 'Verificar integridade' }))
    expect(await screen.findByText('Nenhum problema encontrado.')).toBeInTheDocument()
  })

  it('lista erros e avisos separadamente', async () => {
    installFakeBackend({ farms: [sampleFarm], integrity: { ok: false, checkedAt: '2026-10-11T12:00:00Z', errors: 1, warnings: 1, checked: {}, issues: [
      { severity: 'error', code: 'broken_reference', message: 'Referência quebrada: um registro de animals aponta para farms que não existe.', detail: 'linha 7' },
      { severity: 'warning', code: 'calculation_dataset_removed', message: 'Um cálculo salvo usou a biblioteca Antiga 9, que não está mais instalada.', detail: 'O snapshot mantém os valores.' }] } })
    open()
    fireEvent.click(await screen.findByRole('button', { name: 'Verificar integridade' }))
    const err = await screen.findByRole('alert')
    expect(err).toHaveTextContent('Referência quebrada')
    expect(screen.getByText(/biblioteca Antiga 9/)).toBeInTheDocument()
    expect(screen.getByText((_, el) => el?.tagName === 'P' && /1 erro\(s\) · 1 aviso\(s\)/.test(el.textContent ?? ''))).toBeInTheDocument()
  })

  it('carrega e remove dados de demonstração (remover pede confirmação)', async () => {
    const be = installFakeBackend({ farms: [sampleFarm] })
    open()
    fireEvent.click(await screen.findByRole('button', { name: 'Carregar dados de demonstração' }))
    const rm = await screen.findByRole('button', { name: 'Remover dados de demonstração' })
    fireEvent.click(rm)
    const dlg = await screen.findByRole('dialog', { name: 'Remover dados de demonstração' })
    expect(be.calls.some((c) => c.method === 'RemoveDemoData')).toBe(false)
    fireEvent.click(within(dlg).getByRole('button', { name: 'Remover' }))
    await waitFor(() => expect(be.calls.some((c) => c.method === 'RemoveDemoData')).toBe(true))
    expect(await screen.findByRole('button', { name: 'Carregar dados de demonstração' })).toBeInTheDocument()
  })

  it('lista os atalhos de teclado', async () => {
    installFakeBackend({ farms: [sampleFarm] })
    open()
    const t = await screen.findByRole('table', { name: 'Atalhos de teclado' })
    expect(within(t).getByText('Ir para Rebanho')).toBeInTheDocument()
  })
})
