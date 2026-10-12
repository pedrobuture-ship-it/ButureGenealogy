import { buildManifest, buildMetadata, buildTraits, emptyMeta, slugify, suggestCode, type ColumnMapping } from './importRequest'

describe('importRequest', () => {
  it('gera o id a partir do nome', () => {
    expect(slugify('Avaliação Nelore — 2026/1')).toBe('avaliacao-nelore-2026-1')
    expect(buildManifest({ ...emptyMeta, name: 'Meu Dataset' }).id).toBe('meu-dataset')
    expect(buildManifest({ ...emptyMeta, name: 'X', id: 'custom.id' }).id).toBe('custom.id')
  })

  it('nunca marca a licença como redistribuível por conta própria', () => {
    // sem nome de licença, a caixa é ignorada
    expect(buildManifest({ ...emptyMeta, name: 'X', redistributionAllowed: true }).license.redistributionAllowed).toBe(false)
    expect(buildManifest({ ...emptyMeta, name: 'X' }).license).toEqual({ name: '', url: '', redistributionAllowed: false, commercialUseAllowed: null })
    // com licença nomeada e caixa marcada pelo usuário, sim
    expect(buildManifest({ ...emptyMeta, name: 'X', licenseName: 'CC-BY-4.0', redistributionAllowed: true }).license.redistributionAllowed).toBe(true)
  })

  it('interpreta raças e sexo padrão', () => {
    const m = buildManifest({ ...emptyMeta, name: 'X', breeds: ' NEL, GIR ,, ' })
    expect(m.breeds).toEqual(['NEL', 'GIR'])
    expect(m.data.defaultSex).toBe('male')
    expect(buildManifest({ ...emptyMeta, name: 'X', sexColumn: 'sexo' }).data.defaultSex).toBe('')
  })

  it('só inclui colunas habilitadas e preserva o tipo escolhido', () => {
    const base: ColumnMapping = { column: 'DEP_PD', enabled: true, code: ' PD ', name: 'Peso à desmama', evaluationType: 'DEP', unit: 'kg', category: 'growth', accuracyColumn: 'ACC_PD', percentileColumn: '', existingTraitId: '', percentileSemantics: '', popMean: '', popSd: '' }
    const defs = buildTraits([base, { ...base, column: 'X', enabled: false }, { ...base, column: 'Y', evaluationType: 'EBV', existingTraitId: 'builtin:DEP:WW' }])
    expect(defs).toHaveLength(2)
    expect(defs[0]).toMatchObject({ code: 'PD', evaluationType: 'DEP', accuracyColumn: 'ACC_PD', optimizationDirection: 'INFORMATION_ONLY' })
    expect(defs[1].evaluationType).toBe('EBV')
    expect(defs[1].existingTraitId).toBe('builtin:DEP:WW')
    expect(defs[0].existingTraitId).toBeUndefined()
  })

  it('sugere código de característica a partir da coluna', () => {
    expect(suggestCode('DEP_PD')).toBe('DEP_PD')
    expect(suggestCode('peso (kg)')).toBe('peso__kg_')
    expect(suggestCode('1abc')).toBe('C_1abc')
  })
  it('declara a semântica do percentil e os parâmetros populacionais nos metadados', () => {
    const base = { percentileColumn: 'PCT', percentileSemantics: 'descending' as const, popMean: '10,5', popSd: '2' }
    expect(buildMetadata(base)).toEqual({ percentileSemantics: 'descending', population: { mean: 10.5, sd: 2 } })
    // sem coluna de percentil, a semântica é ignorada; desvio inválido não gera parâmetros
    expect(buildMetadata({ ...base, percentileColumn: '', popMean: '', popSd: '' })).toEqual({})
    expect(buildMetadata({ ...base, popSd: '0' })).toEqual({ percentileSemantics: 'descending' })
    expect(buildMetadata({ ...base, percentileSemantics: '' })).toEqual({ population: { mean: 10.5, sd: 2 } })
  })
})
