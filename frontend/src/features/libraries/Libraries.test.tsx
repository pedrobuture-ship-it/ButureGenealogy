import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import App from '../../App'
import { installFakeBackend, sampleFarm } from '../../test/fakeBackend'

const lib = (over: Record<string, unknown> = {}) => ({
  id: 'l1', packageId: 'synthetic.sires', name: 'Synthetic Sires', version: '1.0.0', origin: 'genlib', species: 'Bos taurus', breeds: ['NEL'], purposes: ['beef'],
  organization: 'Org X', sourceUrl: '', program: 'SYNTHETIC', release: '1.0.0', licenseName: '', licenseUrl: '', licenseStatus: 'unknown',
  redistributionAllowed: false, active: true, isDemo: false, recordCount: 8412, traitCount: 23, fileSha256: 'abc', installedAt: '2026-10-11T00:00:00Z', ...over,
})

const report = (over: Record<string, unknown> = {}) => ({
  records: 8412, rows: 8430, rowsWithoutId: 1, duplicates: 17, values: 100, traitsRecognized: 23, traitsUnrecognized: 3, unrecognizedColumns: ['a', 'b', 'c'],
  invalidValues: 0, withoutAccuracy: 482, accuracyOutOfRange: 0, percentileInvalid: 0, matchedAnimals: 10, unmatchedAnimals: 8402, ambiguousAnimals: 0,
  errors: [], warnings: ['Licença desconhecida: os dados ficam restritos a este computador e não devem ser redistribuídos.'], ...over,
})

const manifest = { schemaVersion: 1, id: 'meu-csv', name: 'Meu CSV', version: '1', species: 'Bos taurus', breeds: ['NEL'], purpose: ['beef'], authors: [], organization: '', sourceUrl: '',
  releaseDate: '', license: { name: '', url: '', redistributionAllowed: false, commercialUseAllowed: null }, evaluation: { program: 'P', release: '1', geneticBase: '' },
  data: { format: 'csv', idColumn: 'registro', nameColumn: 'nome', sexColumn: '', breedColumn: '', defaultSex: 'male', decimalComma: false, accuracyScale: 'fraction' } }

const csvInspection = {
  kind: 'csv', path: '/tmp/avaliacoes.csv', columns: ['registro', 'nome', 'DEP_PD', 'ACC_PD'], sampleRows: [['R1', 'Touro', '5.2', '0.8']], rowCount: 3,
  suggestedDecimalComma: false, suggestedIdColumn: 'registro', suggestedNameColumn: 'nome', manifest: null, traits: [], readme: '', licenseText: '', validation: null,
}

describe('Bibliotecas — instaladas', () => {
  it('mostra o estado vazio e leva à importação', async () => {
    installFakeBackend({ farms: [sampleFarm] })
    window.location.hash = '#/bibliotecas'
    render(<App />)
    expect(await screen.findByText('Nenhuma biblioteca instalada')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Importar biblioteca' }))
    expect(await screen.findByRole('heading', { name: 'Escolha o arquivo' })).toBeInTheDocument()
  })

  it('lista nome, versão, raça, finalidade, fonte, status, licença, registros e traits; licença desconhecida é sinalizada', async () => {
    installFakeBackend({ farms: [sampleFarm], libraries: [lib(), lib({ id: 'l2', name: 'Livre', licenseStatus: 'redistributable', licenseName: 'CC0-1.0', active: false })] })
    window.location.hash = '#/bibliotecas'
    render(<App />)
    const row = (await screen.findByText('Synthetic Sires')).closest('tr')!
    const cells = within(row).getAllByRole('cell')
    expect(cells[1]).toHaveTextContent('1.0.0')
    expect(cells[2]).toHaveTextContent('NEL')
    expect(cells[3]).toHaveTextContent('Corte')
    expect(cells[4]).toHaveTextContent('Org X')
    expect(cells[5]).toHaveTextContent('Ativa')
    expect(cells[6]).toHaveTextContent('Licença desconhecida')
    expect(cells[7]).toHaveTextContent('8.412')
    expect(cells[8]).toHaveTextContent('23')
    const other = screen.getByText('Livre').closest('tr')!
    expect(within(other).getByText('Redistribuível')).toBeInTheDocument()
    expect(within(other).getByText('Inativa')).toBeInTheDocument()
  })

  it('ativa/desativa, abre licença e documentação e remove com confirmação', async () => {
    const be = installFakeBackend({ farms: [sampleFarm], libraries: [lib()] })
    window.location.hash = '#/bibliotecas'
    render(<App />)
    const row = (await screen.findByText('Synthetic Sires')).closest('tr')!

    fireEvent.click(within(row).getByRole('button', { name: 'Desativar' }))
    await waitFor(() => expect(be.calls.some((c) => c.method === 'SetLibraryActive' && c.args[1] === false)).toBe(true))
    expect(await screen.findByText('Inativa')).toBeInTheDocument()

    fireEvent.click(within(screen.getByText('Synthetic Sires').closest('tr')!).getByRole('button', { name: 'Licença' }))
    const dlg = await screen.findByRole('dialog', { name: 'Biblioteca' })
    expect(await within(dlg).findByText('Texto da licença de teste')).toBeInTheDocument()
    expect(within(dlg).getByText(/responsabilidade pela autorização/)).toBeInTheDocument()
    fireEvent.click(within(dlg).getByRole('tab', { name: 'Documentação' }))
    expect(await within(dlg).findByText('# Documentação de teste')).toBeInTheDocument()
    fireEvent.click(within(dlg).getByRole('button', { name: 'Fechar' }))

    fireEvent.click(within(screen.getByText('Synthetic Sires').closest('tr')!).getByRole('button', { name: 'Remover' }))
    const confirm = await screen.findByRole('dialog', { name: 'Remover biblioteca' })
    expect(be.calls.some((c) => c.method === 'RemoveLibrary')).toBe(false)
    fireEvent.click(within(confirm).getByRole('button', { name: 'Remover' }))
    await waitFor(() => expect(be.calls.some((c) => c.method === 'RemoveLibrary')).toBe(true))
    expect(await screen.findByText('Nenhuma biblioteca instalada')).toBeInTheDocument()
  })
})

describe('Bibliotecas — assistente de importação', () => {
  it('CSV: mapeia colunas, mostra a validação no formato combinado e só instala com a declaração de responsabilidade', async () => {
    const be = installFakeBackend({ farms: [sampleFarm], inspection: csvInspection, validation: { report: report(), licenseStatus: 'unknown', duplicate: false, canInstall: true, notice: 'A responsabilidade pela autorização de uso e distribuição de dados importados manualmente é do usuário.', manifest } })
    window.location.hash = '#/bibliotecas'
    render(<App />)
    fireEvent.click(await screen.findByRole('tab', { name: 'Importar biblioteca' }))
    fireEvent.click(screen.getByRole('button', { name: 'Escolher arquivo…' }))

    // step 2
    await screen.findByRole('heading', { name: 'Mapeamento das colunas' })
    const next = screen.getByRole('button', { name: 'Validar importação' })
    expect(next).toBeDisabled() // nothing filled yet

    fireEvent.change(screen.getByLabelText('Nome'), { target: { value: 'Meu CSV' } })
    fireEvent.change(screen.getByLabelText('Versão'), { target: { value: '1' } })
    fireEvent.change(screen.getByLabelText('Programa de avaliação'), { target: { value: 'P' } })
    expect(next).toBeDisabled() // still no mapped column
    fireEvent.click(screen.getByLabelText('Usar a coluna DEP_PD'))
    fireEvent.change(screen.getByLabelText('Coluna de acurácia de DEP_PD'), { target: { value: 'ACC_PD' } })
    fireEvent.change(screen.getByLabelText('Unidade de DEP_PD'), { target: { value: 'kg' } })
    expect(next).toBeEnabled()
    fireEvent.click(next)

    // step 3: the agreed validation format
    const validation = await screen.findByRole('region', { name: 'Validação' })
    expect(within(validation).getByText('registros').closest('li')).toHaveTextContent('8.412 registros')
    expect(within(validation).getByText('características reconhecidas').closest('li')).toHaveTextContent('23 características reconhecidas')
    expect(within(validation).getByText('sem acurácia').closest('li')).toHaveTextContent('482 sem acurácia')
    expect(within(validation).getByText('duplicados').closest('li')).toHaveTextContent('17 duplicados')
    expect(within(validation).getByText('traits não reconhecidos').closest('li')).toHaveTextContent('3 traits não reconhecidos')
    expect(screen.getByText('Licença desconhecida', { selector: '.badge' })).toBeInTheDocument()

    // the request sent for validation carries exactly what the user mapped
    const sent = be.calls.find((c) => c.method === 'ValidateImport')!.args[0] as { traits: { code: string; column: string; accuracyColumn: string; unit: string; evaluationType: string }[]; manifest: { license: { redistributionAllowed: boolean } } }
    expect(sent.traits).toHaveLength(1)
    expect(sent.traits[0]).toMatchObject({ column: 'DEP_PD', accuracyColumn: 'ACC_PD', unit: 'kg', evaluationType: 'DEP' })
    expect(sent.manifest.license.redistributionAllowed).toBe(false)

    const install = screen.getByRole('button', { name: 'Instalar biblioteca' })
    expect(install).toBeDisabled()
    expect(screen.getAllByText(/responsabilidade pela autorização de uso e distribuição de dados importados manualmente é do usuário/).length).toBeGreaterThan(0)
    fireEvent.click(screen.getByLabelText('Declaro que tenho autorização para usar estes dados.'))
    expect(install).toBeEnabled()

    // choosing to create animals requires a farm
    fireEvent.click(screen.getByLabelText(/Cadastrar como novos animais/))
    expect(install).toBeDisabled()
    fireEvent.change(screen.getByLabelText('Propriedade para os novos animais'), { target: { value: 'f1' } })
    expect(install).toBeEnabled()
    fireEvent.click(install)

    expect(await screen.findByRole('heading', { name: 'Biblioteca instalada' })).toBeInTheDocument()
    expect(screen.getByText(/6 avaliações importadas/)).toBeInTheDocument()
    const installReq = be.calls.find((c) => c.method === 'InstallLibrary')!.args[0] as { acceptResponsibility: boolean; unmatchedPolicy: string; farmId: string }
    expect(installReq).toMatchObject({ acceptResponsibility: true, unmatchedPolicy: 'create', farmId: 'f1' })
  })

  it('bloqueia a instalação quando a validação encontra erros', async () => {
    installFakeBackend({ farms: [sampleFarm], inspection: csvInspection, validation: { report: report({ errors: ['A coluna "DEP_PD" não existe no arquivo.'] }), licenseStatus: 'unknown', duplicate: false, canInstall: false, notice: '', manifest } })
    window.location.hash = '#/bibliotecas'
    render(<App />)
    fireEvent.click(await screen.findByRole('tab', { name: 'Importar biblioteca' }))
    fireEvent.click(screen.getByRole('button', { name: 'Escolher arquivo…' }))
    await screen.findByRole('heading', { name: 'Mapeamento das colunas' })
    fireEvent.change(screen.getByLabelText('Nome'), { target: { value: 'X' } })
    fireEvent.change(screen.getByLabelText('Versão'), { target: { value: '1' } })
    fireEvent.change(screen.getByLabelText('Programa de avaliação'), { target: { value: 'P' } })
    fireEvent.click(screen.getByLabelText('Usar a coluna DEP_PD'))
    fireEvent.click(screen.getByRole('button', { name: 'Validar importação' }))
    expect(await screen.findByText('A coluna "DEP_PD" não existe no arquivo.')).toBeInTheDocument()
    fireEvent.click(screen.getByLabelText('Declaro que tenho autorização para usar estes dados.'))
    expect(screen.getByRole('button', { name: 'Instalar biblioteca' })).toBeDisabled()
  })

  it('.genlib vai direto para a validação, sem mapeamento', async () => {
    installFakeBackend({
      farms: [sampleFarm],
      inspection: { ...csvInspection, kind: 'genlib', path: '/tmp/x.genlib', manifest: { ...manifest, license: { name: 'CC0-1.0', url: 'https://x', redistributionAllowed: true, commercialUseAllowed: null } },
        validation: { report: report({ warnings: [] }), licenseStatus: 'redistributable', duplicate: false, canInstall: true, notice: 'n', manifest: { ...manifest, license: { name: 'CC0-1.0', url: 'https://x', redistributionAllowed: true, commercialUseAllowed: null } } } },
    })
    window.location.hash = '#/bibliotecas'
    render(<App />)
    fireEvent.click(await screen.findByRole('tab', { name: 'Importar biblioteca' }))
    fireEvent.click(screen.getByRole('button', { name: 'Escolher arquivo…' }))
    expect(await screen.findByRole('region', { name: 'Validação' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'Mapeamento das colunas' })).not.toBeInTheDocument()
    expect(screen.getByText('Redistribuível')).toBeInTheDocument()
  })
})

describe('Bibliotecas — catálogo de características', () => {
  it('lista o catálogo e explica os tipos', async () => {
    installFakeBackend({ farms: [sampleFarm] })
    window.location.hash = '#/bibliotecas'
    render(<App />)
    fireEvent.click(await screen.findByRole('tab', { name: 'Características' }))
    expect(await screen.findByText('BODY_WEIGHT')).toBeInTheDocument()
    expect(await screen.findByText('O que significa cada tipo de valor?')).toBeInTheDocument()
  })
})
