// Mirrors the Go domain/service DTOs. Keep in sync with internal/domain/entities.go.
export type Sex = 'male' | 'female'
export type Category = 'calf' | 'heifer' | 'cow' | 'steer' | 'bull' | 'other'
export type AnimalStatus = 'active' | 'archived' | 'sold' | 'deceased'
export type EvaluationType = 'DEP' | 'EPD' | 'EBV' | 'PTA' | 'INDEX'

export interface Farm {
  id: string
  name: string
  ownerName: string
  city: string
  state: string
  notes: string
  isDemo: boolean
  createdAt: string
  updatedAt: string
}
export interface FarmInput {
  name: string
  ownerName: string
  city: string
  state: string
  notes: string
}

export interface Breed {
  id: string
  code: string
  name: string
  species: string
  purpose: 'beef' | 'dairy' | 'dual'
  active: boolean
}

export interface Animal {
  id: string
  farmId: string
  registration: string
  name: string
  sex: Sex
  breedId: string
  birthDate: string
  category: Category
  status: AnimalStatus
  fatherId: string
  motherId: string
  notes: string
  isDemo: boolean
  createdAt: string
  updatedAt: string
}
export type AnimalInput = Omit<Animal, 'id' | 'isDemo' | 'createdAt' | 'updatedAt'>

export interface AnimalQuery {
  farmId: string
  breedId: string
  sex: string
  category: string
  status: string
  includeArchived: boolean
  search: string
  minAgeMonths: number | null
  maxAgeMonths: number | null
  sortBy: string
  sortDesc: boolean
  page: number
  pageSize: number
}
export interface AnimalPage {
  items: Animal[]
  total: number
  page: number
}

export interface AnimalRef {
  id: string
  name: string
  registration: string
  sex: Sex
  breedName: string
}
export interface AnimalDetail {
  animal: Animal
  breedName: string
  farmName: string
  father: AnimalRef | null
  mother: AnimalRef | null
  offspring: AnimalRef[]
  semenDoses: number
}

export interface Measurement {
  id: string
  animalId: string
  traitCode: string
  value: number
  unit: string
  measuredAt: string
  source: string
  notes: string
  isDemo: boolean
  createdAt: string
}
export interface MeasurementInput {
  animalId: string
  traitCode: string
  value: number
  unit: string
  measuredAt: string
  source: string
  notes: string
}

export interface GeneticEvaluation {
  id: string
  animalId: string
  datasetId: string
  evaluationProgram: string
  evaluationRelease: string
  breedCode: string
  traitCode: string
  evaluationType: EvaluationType
  value: number
  unit: string
  accuracy: number | null
  percentile: number | null
  geneticBase: string
  sourceReference: string
  evaluatedAt: string
  isDemo: boolean
  createdAt: string
}

// ---- Pedigree (internal/genetics, internal/service/pedigree.go)
export interface TreeNode {
  id: string
  name: string
  registration: string
  sex: Sex | ''
  breedName: string
  birthDate: string
  generation: number
  role: 'self' | 'sire' | 'dam'
  known: boolean
  childId: string
  childName: string
  hasMore: boolean
  sire: TreeNode | null
  dam: TreeNode | null
}
export interface GenerationStat { generation: number; expected: number; known: number; percent: number }
export interface RepeatedAncestor { id: string; name: string; occurrences: string[] }
export interface PedigreeIssue { severity: 'error' | 'warning'; code: string; animalId: string; message: string }
export interface CoefficientResult { status: 'not_calculated' | 'calculated'; value: number | null; method: string; message: string }
export interface PedigreeView {
  animalId: string
  generations: number
  tree: TreeNode | null
  depth: number
  stats: GenerationStat[]
  repeated: RepeatedAncestor[]
  issues: PedigreeIssue[]
  inbreeding: CoefficientResult
}
export interface MinimalAncestorInput { childId: string; role: 'sire' | 'dam'; name: string; registration: string; breedId: string }

// ---- Traits (internal/domain/trait.go, internal/service/trait.go)
export type TraitCategory =
  'growth' | 'reproduction' | 'maternal' | 'carcass' | 'dairy' | 'conformation' | 'health' | 'temperament' | 'efficiency' | 'other'
export type ValueKind = 'PHENOTYPE' | EvaluationType
export type OptimizationDirection = 'MAXIMIZE' | 'MINIMIZE' | 'TARGET_RANGE' | 'INFORMATION_ONLY'

export interface Trait {
  id: string
  code: string
  name: string
  description: string
  category: TraitCategory
  purpose: 'beef' | 'dairy' | 'dual' | 'any'
  unit: string
  dataType: 'number' | 'integer' | 'score'
  evaluationType: ValueKind
  optimizationDirection: OptimizationDirection
  idealMin: number | null
  idealMax: number | null
  higherIsBetter: boolean | null
  sourceDataset: string
  metadata: string
  isBuiltin: boolean
  createdAt: string
}
export interface TraitInput {
  code: string
  name: string
  description: string
  category: string
  purpose: string
  unit: string
  dataType: string
  evaluationType: string
  optimizationDirection: string
  idealMin: number | null
  idealMax: number | null
  higherIsBetter: boolean | null
}
export interface Provenance {
  kindCode: ValueKind
  kindLabel: string
  genetic: boolean
  traitCode: string
  traitName: string
  datasetId: string
  datasetName: string
  datasetVersion: string
  program: string
  release: string
  geneticBase: string
  sourceReference: string
  accuracy: number | null
  source: string
  date: string
  isDemo: boolean
}
export interface TraitItem {
  id: string
  source: 'measurement' | 'evaluation'
  traitCode: string
  traitName: string
  defined: boolean
  category: TraitCategory
  valueKind: ValueKind
  unit: string
  value: number
  accuracy: number | null
  percentile: number | null
  notes: string
  optimizationDirection: OptimizationDirection
  provenance: Provenance
}
export interface TraitGroup { category: TraitCategory; items: TraitItem[] }
export interface ValueKindInfo { kind: ValueKind; label: string; description: string; genetic: boolean }

// ---- Libraries (internal/datasets, internal/service/library.go)
export type LicenseStatus = 'unknown' | 'no_redistribute' | 'redistributable'

export interface LibrarySummary {
  id: string
  packageId: string
  name: string
  version: string
  origin: 'genlib' | 'csv' | 'json'
  species: string
  breeds: string[]
  purposes: string[]
  organization: string
  sourceUrl: string
  program: string
  release: string
  licenseName: string
  licenseUrl: string
  licenseStatus: LicenseStatus
  redistributionAllowed: boolean
  active: boolean
  isDemo: boolean
  recordCount: number
  traitCount: number
  fileSha256: string
  installedAt: string
}
export interface LibraryDetail extends LibrarySummary {
  manifest: Record<string, unknown>
  readme: string
  licenseText: string
  traits: Trait[]
  usedByCalculations: number
  responsibilityNotice: string
}

export interface LibraryLicense {
  name: string
  url: string
  redistributionAllowed: boolean
  commercialUseAllowed: boolean | null
}
export interface LibraryManifest {
  schemaVersion: number
  id: string
  name: string
  version: string
  species: string
  breeds: string[]
  purpose: string[]
  authors: string[]
  organization: string
  sourceUrl: string
  releaseDate: string
  license: LibraryLicense
  evaluation: { program: string; release: string; geneticBase: string }
  data: {
    format: string
    idColumn: string
    nameColumn: string
    sexColumn: string
    breedColumn: string
    defaultSex: string
    decimalComma: boolean
    accuracyScale: 'fraction' | 'percent'
  }
}
export interface TraitDef {
  code: string
  name: string
  description: string
  category: string
  purpose: string
  unit: string
  dataType: string
  evaluationType: string
  optimizationDirection: string
  idealMin: number | null
  idealMax: number | null
  higherIsBetter: boolean | null
  metadata: Record<string, unknown>
  column: string
  accuracyColumn: string
  percentileColumn: string
  existingTraitId?: string
}
export interface ImportReport {
  records: number
  rows: number
  rowsWithoutId: number
  duplicates: number
  values: number
  traitsRecognized: number
  traitsUnrecognized: number
  unrecognizedColumns: string[]
  invalidValues: number
  withoutAccuracy: number
  accuracyOutOfRange: number
  percentileInvalid: number
  matchedAnimals: number
  unmatchedAnimals: number
  ambiguousAnimals: number
  errors: string[]
  warnings: string[]
}
export interface ImportValidation {
  report: ImportReport
  licenseStatus: LicenseStatus
  duplicate: boolean
  canInstall: boolean
  notice: string
  manifest: LibraryManifest
}
export interface ImportInspection {
  kind: 'genlib' | 'csv' | 'json'
  path: string
  columns: string[]
  sampleRows: string[][]
  rowCount: number
  suggestedDecimalComma: boolean
  suggestedIdColumn: string
  suggestedNameColumn: string
  manifest: LibraryManifest | null
  traits: TraitDef[]
  readme: string
  licenseText: string
  validation: ImportValidation | null
}
export interface ImportRequest {
  path: string
  manifest: LibraryManifest
  traits: TraitDef[]
  licenseText: string
  unmatchedPolicy: 'skip' | 'create'
  farmId: string
  acceptResponsibility: boolean
}
export interface InstallResult {
  library: LibrarySummary
  report: ImportReport
  evaluationsInserted: number
  animalsCreated: number
  recordsSkipped: number
}

// ---- Semen (internal/service/semen.go)
export type StockStatus = 'in_stock' | 'depleted' | 'no_semen'
export interface BankRow {
  id: string
  name: string
  registration: string
  breedName: string
  animalStatus: AnimalStatus
  doses: number
  lots: number
  suppliers: string[]
  unitCost: number | null
  stockStatus: StockStatus
  isDemo: boolean
}
export interface BankPage { items: BankRow[]; total: number; page: number; totalDoses: number; suppliers: string[] }
export interface BankQuery {
  search: string
  breedId: string
  supplier: string
  stockFilter: 'with_lots' | 'in_stock' | 'all'
  includeArchived: boolean
  sortBy: string
  sortDesc: boolean
  page: number
  pageSize: number
}
export interface SemenLot {
  id: string
  bullId: string
  supplier: string
  batch: string
  quantity: number
  unitCost: number | null
  semenType: string
  notes: string
  isDemo: boolean
  createdAt: string
  updatedAt: string
}
export interface SemenMovement {
  id: string
  inventoryId: string
  bullId: string
  type: 'entry' | 'withdrawal' | 'adjustment'
  delta: number
  balanceAfter: number
  reason: string
  note: string
  matingId: string
  isDemo: boolean
  createdAt: string
}
export interface SemenType { code: string; name: string; isBuiltin: boolean }
export interface BullSheet {
  detail: AnimalDetail
  lots: SemenLot[]
  totalDoses: number
  unitCost: number | null
  movements: SemenMovement[]
  stockStatus: StockStatus
}
export interface SemenInput {
  bullId: string
  supplier: string
  batch: string
  quantity: number
  unitCost: number | null
  semenType: string
  notes: string
  reason: string
}
export interface LotInfoInput { supplier: string; batch: string; unitCost: number | null; semenType: string; notes: string }
export interface MovementInput { inventoryId: string; quantity: number; reason: string; note: string; matingId: string }

// ---- Backup and integrity (internal/backup)
export interface BackupLibraryEntry {
  id: string
  packageId: string
  name: string
  version: string
  fileName: string
  sha256: string
  licenseStatus: string
  isDemo: boolean
  isOfficial: boolean
  included: boolean
}
export interface BackupManifest {
  formatVersion: number
  createdAt: string
  appVersion: string
  engineVersion: string
  schemaVersion: number
  farms: string[]
  counts: Record<string, number>
  files: { path: string; size: number; sha256: string }[]
  libraries: BackupLibraryEntry[]
}
export interface BackupPreview {
  manifest: BackupManifest
  upgradedFrom: number
  schemaVersion: number
  path: string
}
export interface RestoreResult { safetyBackup: string; counts: Record<string, number>; missingLibraries: string[] }
export interface IntegrityIssue { severity: 'error' | 'warning'; code: string; message: string; detail: string }
export interface IntegrityReport {
  ok: boolean
  checkedAt: string
  errors: number
  warnings: number
  issues: IntegrityIssue[]
  checked: Record<string, number>
}
