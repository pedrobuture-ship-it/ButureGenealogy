// Sections of the "Cálculos" page, used by "Ver cálculo" to link to the matching explanation.
export const docSection = {
  score: 'score',
  weights: 'pesos',
  normalization: 'normalizacao',
  maximize: 'maximize',
  minimize: 'minimize',
  targetRange: 'target-range',
  accuracy: 'acuracia',
  complementarity: 'complementaridade',
  pedigree: 'pedigree',
  relationship: 'parentesco',
  progeny: 'progenie',
  datasets: 'datasets',
  reproducibility: 'reproducibilidade',
} as const

export const docPath = (section: string) => `/calculos/${section}`

const directionDoc: Record<string, string> = { MAXIMIZE: docSection.maximize, MINIMIZE: docSection.minimize, TARGET_RANGE: docSection.targetRange }
export const directionSection = (d: string) => directionDoc[d] ?? docSection.normalization
