// Mirrors internal/engine. The frontend only renders these numbers; it never recomputes them.
export type Direction = 'MAXIMIZE' | 'MINIMIZE' | 'TARGET_RANGE' | 'INFORMATION_ONLY'
export type CalcStatus = 'calculated' | 'not_calculable' | 'not_applied' | 'excluded'

export interface TraitWeight {
  traitCode: string
  evaluationType: string
  weight: number
  direction: '' | 'MAXIMIZE' | 'MINIMIZE' | 'TARGET_RANGE'
  idealMin: number | null
  idealMax: number | null
  targetTolerance: number | null
  required: boolean
}
export interface EngineFilters {
  requireStock: boolean
  minDoses: number
  maxUnitCost: number | null
  minAccuracy: number | null
  excludeDirectAncestors: boolean
  ancestorGenerations: number
}
export interface Profile {
  id: string
  name: string
  isDemo: boolean
  demoNotice: string
  description: string
  traits: TraitWeight[]
  filters: EngineFilters
  normalization: 'auto' | 'percentile' | 'zscore'
  minCoverage: number
  complementarity: { maxBonusPoints: number }
  penalties: { lowAccuracy: { threshold: number; maxPoints: number } }
}

export interface Basis {
  traitCode: string
  evaluationType: string
  program: string
  release: string
  geneticBase: string
  breedCode: string
  unit: string
}
export interface Source {
  evaluationId: string
  datasetId: string
  datasetName: string
  datasetVersion: string
  program: string
  release: string
  geneticBase: string
  isDemo: boolean
  parameter: string
}
export interface EngineAnimal {
  id: string
  name: string
  registration: string
  sex: string
  status: string
  breedName: string
  isDemo: boolean
}
export interface ComponentResult {
  name: string
  traitCode: string
  evaluationType: string
  status: CalcStatus
  direction: string
  inputValue: number | null
  damValue: number | null
  unit: string
  accuracy: number | null
  percentile: number | null
  score: number | null
  method: string
  transformation: string
  formula: string
  rawWeight: number
  weight: number
  contribution: number | null
  basis: Basis | null
  source: Source
  explanation: string
  warnings: string[]
}
export interface ComplementarityItem {
  traitCode: string
  name: string
  status: CalcStatus
  bullScore: number | null
  damScore: number | null
  gain: number | null
  weight: number
  points: number
  explanation: string
  source: Source
}
export interface ComplementarityResult {
  status: CalcStatus
  points: number
  maxBonus: number
  formula: string
  items: ComplementarityItem[]
  source: Source
  explanation: string
  warnings: string[]
}
export interface PenaltyResult {
  name: string
  status: CalcStatus
  points: number
  formula: string
  source: Source
  explanation: string
  warnings: string[]
}
export interface FinalScore {
  base: number
  complementarity: number
  penalties: number
  unclamped: number
  final: number
  clamped: boolean
  coverage: number
  formula: string
}
export interface CandidateResult {
  bull: EngineAnimal
  stock: { doses: number; unitCost: number | null }
  status: CalcStatus
  rank: number
  tied: boolean
  exclusions: { filter: string; reason: string }[]
  reasons: string[]
  components: ComponentResult[]
  complementarity: ComplementarityResult
  penalties: PenaltyResult[]
  final: FinalScore | null
  warnings: string[]
  pedigreeNotes: string[]
}
export interface ReferenceBasis {
  traitCode: string
  evaluationType: string
  basis: Basis | null
  candidates: number
  eligible: number
  damMatches: boolean
  reason: string
  alternatives: { basis: Basis; candidates: number }[]
}
export interface DatasetInfo {
  id: string
  packageId: string
  name: string
  version: string
  program: string
  release: string
  sha256: string
  isDemo: boolean
  origin: string
  licenseStatus: string
}
export interface RunOutput {
  engineVersion: string
  createdAt: string
  female: EngineAnimal
  profile: Profile
  bases: ReferenceBasis[]
  results: CandidateResult[]
  top: string[]
  datasets: DatasetInfo[]
  errors: string[]
  warnings: string[]
}
export interface EngineEvaluation {
  id: string
  basis: Basis
  value: number
  accuracy: number | null
  percentile: number | null
  datasetName: string
  datasetVersion: string
  isDemo: boolean
  traitName: string
}
export interface PairSnapshot {
  snapshotVersion: number
  engineVersion: string
  createdAt: string
  female: EngineAnimal
  femaleEvals: EngineEvaluation[]
  bull: EngineAnimal
  bullEvals: EngineEvaluation[]
  profile: Profile
  bases: ReferenceBasis[]
  datasets: DatasetInfo[]
  result: CandidateResult
  candidates: number
  runWarnings: string[]
  isDemo: boolean
}
export interface RunRequest { femaleId: string; bullIds: string[]; profile: Profile; save: boolean }
export interface RunResponse {
  runId: string
  saved: boolean
  matingIds: Record<string, string>
  output: RunOutput
  snapshots: Record<string, PairSnapshot>
}
export interface CalculationView {
  id: string
  runId: string
  femaleId: string
  bullId: string
  status: string
  score: number | null
  engineVersion: string
  createdAt: string
  isDemo: boolean
  snapshot: PairSnapshot
}
export interface CalculationSummary {
  id: string
  runId: string
  bullId: string
  bullName: string
  profileName: string
  status: string
  score: number | null
  rank: number
  engineVersion: string
  createdAt: string
  isDemo: boolean
}
