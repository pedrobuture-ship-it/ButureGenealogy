import type { CandidateResult, ComponentResult, PairSnapshot, Profile, RunResponse } from '../types/engine'

export const demoNotice = 'DEMO — NÃO REPRESENTA RECOMENDAÇÃO ZOOTÉCNICA VALIDADA'

export const profileOf = (over: Partial<Profile> = {}): Profile => ({
  id: 'builtin:custom', name: 'Personalizada', isDemo: false, demoNotice: '', description: '', traits: [],
  filters: { requireStock: true, minDoses: 0, maxUnitCost: null, minAccuracy: null, excludeDirectAncestors: false, ancestorGenerations: 3 },
  normalization: 'auto', minCoverage: 1, complementarity: { maxBonusPoints: 0 }, penalties: { lowAccuracy: { threshold: 0, maxPoints: 0 } }, ...over,
})

const source = { evaluationId: 'e1', datasetId: 'd1', datasetName: 'Nelore Example', datasetVersion: '2026.1', program: 'Programa X', release: '2026.1', geneticBase: 'base 2020', isDemo: false, parameter: 'perfil "Personalizada": peso 30,00' }

export const component = (code: string, name: string, score: number, weight: number, contribution: number, over: Partial<ComponentResult> = {}): ComponentResult => ({
  name, traitCode: code, evaluationType: 'DEP', status: 'calculated', direction: 'MAXIMIZE', inputValue: 6.1, damValue: 2.2, unit: 'kg', accuracy: 0.81, percentile: score,
  score, method: 'percentile', transformation: `percentil ${score},00 → ${score},00 (percentil ascendente)`, formula: 'score = percentil', rawWeight: weight * 100, weight, contribution,
  basis: { traitCode: code, evaluationType: 'DEP', program: 'Programa X', release: '2026.1', geneticBase: 'base 2020', breedCode: 'NEL', unit: 'kg' },
  source, explanation: 'ok', warnings: [], ...over,
})

/** The numbers of the product spec: 64,00 + 4,10 − 2,30 = 65,80. */
export const specSnapshot = (over: Partial<PairSnapshot> = {}): PairSnapshot => ({
  snapshotVersion: 1, engineVersion: '0.1.0', createdAt: '2026-10-11T12:00:00Z',
  female: { id: 'f', name: 'Matriz', registration: '0472', sex: 'female', status: 'active', breedName: 'Nelore', isDemo: false },
  femaleEvals: [{ id: 'fe', basis: { traitCode: 'FERT', evaluationType: 'DEP', program: 'Programa X', release: '2026.1', geneticBase: 'base 2020', breedCode: 'NEL', unit: 'kg' }, value: 2.2, accuracy: 0.5, percentile: 55, datasetName: 'Nelore Example', datasetVersion: '2026.1', isDemo: false, traitName: 'Fertilidade' }],
  bull: { id: 'b', name: 'Touro A', registration: '', sex: 'male', status: 'active', breedName: 'Nelore', isDemo: false },
  bullEvals: [], profile: profileOf(), bases: [], candidates: 4,
  datasets: [{ id: 'd1', packageId: 'nelore.example', name: 'Nelore Example', version: '2026.1', program: 'Programa X', release: '2026.1', sha256: 'abc', isDemo: false, origin: 'genlib', licenseStatus: 'redistributable' }],
  runWarnings: [], isDemo: false,
  result: {
    bull: { id: 'b', name: 'Touro A', registration: '', sex: 'male', status: 'active', breedName: 'Nelore', isDemo: false }, stock: { doses: 10, unitCost: 45 }, status: 'calculated', rank: 1, tied: false,
    exclusions: [], reasons: [],
    components: [component('FERT', 'Fertilidade (DEP)', 82, 0.3, 24.6), component('CARC', 'Carcaça (DEP)', 91, 0.25, 22.75), component('MAT', 'Maternal (DEP)', 76, 0.2, 15.2)],
    complementarity: { status: 'calculated', points: 4.1, maxBonus: 10, formula: 'f', source: { ...source, parameter: 'perfil: bônus máximo = 10' }, explanation: 'Bônus pela melhora do touro sobre a matriz.', warnings: [],
      items: [{ traitCode: 'FERT', name: 'Fertilidade (DEP)', status: 'calculated', bullScore: 82, damScore: 55, gain: 0.27, weight: 0.3, points: 0.81, explanation: '', source }] },
    penalties: [
      { name: 'Baixa acurácia', status: 'calculated', points: 2.3, formula: 'f', source: { ...source, parameter: 'perfil: limiar 0,60' }, explanation: 'Acurácias abaixo de 60% reduzem o score em 2,30 ponto(s).', warnings: [] },
      { name: 'Parentesco / endogamia', status: 'not_applied', points: 0, formula: '', source: { ...source, parameter: '' }, explanation: 'Planejado — ainda não utilizado no score do MVP.', warnings: [] },
    ],
    final: { base: 64, complementarity: 4.1, penalties: 2.3, unclamped: 65.8, final: 65.8, clamped: false, coverage: 1, formula: 'final = limitar(base + complementaridade − penalidades, 0, 100)' },
    warnings: [], pedigreeNotes: [],
  } satisfies CandidateResult,
  ...over,
})

export const notCalculableSnapshot = (): PairSnapshot => {
  const s = specSnapshot()
  s.result = { ...s.result, status: 'not_calculable', rank: 0, final: null, reasons: ['Apenas 80,00% do peso do perfil é calculável (mínimo exigido: 100,00%).'],
    components: [component('FERT', 'Fertilidade (DEP)', 82, 0, 0, { status: 'not_calculable', score: null, contribution: null, weight: 0, explanation: 'Touro A tem DEP de FERT apenas em base diferente.', warnings: ['⚠ Comparação genética não suportada entre essas bases.'] })] }
  return s
}

export const runResponse = (): RunResponse => {
  const a = specSnapshot()
  const b = specSnapshot()
  b.bull = { ...a.bull, id: 'b2', name: 'Touro B' }
  b.result = { ...b.result, bull: b.bull, rank: 2, final: { ...b.result.final!, final: 60.5 } }
  const g = notCalculableSnapshot()
  g.bull = { ...a.bull, id: 'b3', name: 'Touro Gama' }
  g.result = { ...g.result, bull: g.bull, warnings: ['FERT: ⚠ Comparação genética não suportada entre essas bases.'] }
  const x = specSnapshot()
  x.result = { ...x.result, bull: { ...a.bull, id: 'b4', name: 'Touro Sem Estoque' }, status: 'excluded', final: null, components: [], exclusions: [{ filter: 'estoque', reason: 'Estoque insuficiente: 0 dose(s), mínimo exigido 1.' }], rank: 0 }
  return {
    runId: 'run1', saved: true, matingIds: { b: 'm1', b2: 'm2', b3: 'm3' },
    output: { engineVersion: '0.1.0', createdAt: '2026-10-11T12:00:00Z', female: a.female, profile: profileOf({ name: 'Personalizada' }), bases: [], top: ['b', 'b2'], datasets: a.datasets, errors: [], warnings: [],
      results: [a.result, b.result, g.result, x.result] },
    snapshots: { b: a, b2: b, b3: g },
  }
}
