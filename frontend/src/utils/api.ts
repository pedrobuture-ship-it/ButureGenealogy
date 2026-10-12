// Single gateway to the Go backend. Pages never call window.go directly,
// and no business rule lives on the frontend: it only renders what Go returns.
import type { AppInfo, DbHealth } from '../types/app'
import type { CalculationSummary, CalculationView, Profile, RunRequest, RunResponse } from '../types/engine'
import type {
  Animal, AnimalDetail, BackupManifest, BackupPreview, IntegrityReport, RestoreResult, AnimalInput, AnimalPage, AnimalQuery, BankPage, BankQuery, Breed, BullSheet, Farm, FarmInput,
  GeneticEvaluation, ImportInspection, ImportRequest, ImportValidation, InstallResult, LibraryDetail, LibrarySummary,
  LotInfoInput, Measurement, MeasurementInput, MinimalAncestorInput, MovementInput, PedigreeIssue, PedigreeView,
  SemenInput, SemenLot, SemenMovement, SemenType, Trait, TraitGroup, TraitInput, ValueKindInfo,
} from '../types/domain'

type Method = (...args: unknown[]) => Promise<unknown>

declare global {
  interface Window {
    go?: { main?: { App?: Record<string, Method> } }
  }
}

/** Structured error produced by the Go layer (see internal/service/wire.go). */
export class ApiError extends Error {
  code: string
  fields: Record<string, string>
  constructor(code: string, message: string, fields: Record<string, string> = {}) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.fields = fields
  }
}

export class BackendUnavailableError extends ApiError {
  constructor() {
    super('unavailable', 'O backend do aplicativo não está disponível (executando fora do Wails).')
    this.name = 'BackendUnavailableError'
  }
}

function toApiError(e: unknown): ApiError {
  if (e instanceof ApiError) return e
  const text = e instanceof Error ? e.message : typeof e === 'string' ? e : ''
  try {
    const w = JSON.parse(text) as { code?: string; message?: string; fields?: Record<string, string> }
    if (w && typeof w.message === 'string') return new ApiError(w.code ?? 'internal', w.message, w.fields ?? {})
  } catch {
    /* not a wire error */
  }
  return new ApiError('internal', text || 'Ocorreu um erro inesperado.')
}

async function call<T>(method: string, ...args: unknown[]): Promise<T> {
  const fn = window.go?.main?.App?.[method]
  if (!fn) throw new BackendUnavailableError()
  try {
    return (await fn(...args)) as T
  } catch (e) {
    throw toApiError(e)
  }
}

export const api = {
  getAppInfo: () => call<AppInfo>('GetAppInfo'),
  healthCheck: () => call<DbHealth>('HealthCheck'),

  listFarms: () => call<Farm[]>('ListFarms'),
  createFarm: (in_: FarmInput) => call<Farm>('CreateFarm', in_),
  listBreeds: () => call<Breed[]>('ListBreeds'),

  listAnimals: (q: Partial<AnimalQuery>) => call<AnimalPage>('ListAnimals', q),
  getAnimal: (id: string) => call<Animal>('GetAnimal', id),
  getAnimalDetail: (id: string) => call<AnimalDetail>('GetAnimalDetail', id),
  createAnimal: (in_: AnimalInput) => call<Animal>('CreateAnimal', in_),
  updateAnimal: (id: string, in_: AnimalInput) => call<Animal>('UpdateAnimal', id, in_),
  archiveAnimal: (id: string) => call<void>('ArchiveAnimal', id),
  restoreAnimal: (id: string) => call<void>('RestoreAnimal', id),

  listMeasurements: (animalId: string) => call<Measurement[]>('ListMeasurements', animalId),
  addMeasurement: (in_: MeasurementInput) => call<Measurement>('AddMeasurement', in_),
  deleteMeasurement: (id: string) => call<void>('DeleteMeasurement', id),
  listEvaluations: (animalId: string) => call<GeneticEvaluation[]>('ListEvaluations', animalId),

  getPedigree: (animalId: string, generations: number) => call<PedigreeView>('GetPedigree', animalId, generations),
  auditPedigrees: (farmId: string) => call<PedigreeIssue[]>('AuditPedigrees', farmId),
  createMinimalAncestor: (in_: MinimalAncestorInput) => call<Animal>('CreateMinimalAncestor', in_),

  listTraits: (q: { kind?: string; category?: string; search?: string } = {}) => call<Trait[]>('ListTraits', q),
  createTrait: (in_: TraitInput) => call<Trait>('CreateTrait', in_),
  listValueKinds: () => call<ValueKindInfo[]>('ListValueKinds'),
  getAnimalTraits: (animalId: string) => call<TraitGroup[]>('GetAnimalTraits', animalId),

  pickLibraryFile: () => call<string>('PickLibraryFile'),
  listLibraries: () => call<LibrarySummary[]>('ListLibraries'),
  getLibrary: (id: string) => call<LibraryDetail>('GetLibrary', id),
  setLibraryActive: (id: string, active: boolean) => call<void>('SetLibraryActive', id, active),
  removeLibrary: (id: string) => call<void>('RemoveLibrary', id),
  inspectImport: (path: string) => call<ImportInspection>('InspectImport', path),
  validateImport: (req: ImportRequest) => call<ImportValidation>('ValidateImport', req),
  installLibrary: (req: ImportRequest) => call<InstallResult>('InstallLibrary', req),

  semenBank: (q: Partial<BankQuery>) => call<BankPage>('SemenBank', q),
  getBullSheet: (bullId: string) => call<BullSheet>('GetBullSheet', bullId),
  listSemenTypes: () => call<SemenType[]>('ListSemenTypes'),
  createSemenType: (name: string) => call<SemenType>('CreateSemenType', name),
  addSemenLot: (in_: SemenInput) => call<SemenLot>('AddSemenLot', in_),
  updateSemenLot: (id: string, in_: LotInfoInput) => call<SemenLot>('UpdateSemenLot', id, in_),
  receiveSemen: (in_: MovementInput) => call<SemenMovement>('ReceiveSemen', in_),
  withdrawSemen: (in_: MovementInput) => call<SemenMovement>('WithdrawSemen', in_),
  adjustSemen: (in_: MovementInput) => call<SemenMovement>('AdjustSemen', in_),

  runMating: (req: RunRequest) => call<RunResponse>('RunMating', req),
  getCalculation: (matingId: string) => call<CalculationView>('GetCalculation', matingId),
  listCalculations: (femaleId: string) => call<CalculationSummary[]>('ListCalculations', femaleId),
  setMatingStatus: (matingId: string, status: string) => call<void>('SetMatingStatus', matingId, status),
  listMatingProfiles: () => call<Profile[]>('ListMatingProfiles'),
  saveMatingProfile: (p: Profile) => call<Profile>('SaveMatingProfile', p),
  deleteMatingProfile: (id: string) => call<void>('DeleteMatingProfile', id),

  engineVersion: () => call<string>('EngineVersion'),

  suggestedBackupName: () => call<string>('SuggestedBackupName'),
  pickBackupSavePath: (defaultName: string) => call<string>('PickBackupSavePath', defaultName),
  pickBackupOpenPath: () => call<string>('PickBackupOpenPath'),
  createBackup: (path: string) => call<BackupManifest>('CreateBackup', path),
  inspectBackup: (path: string) => call<BackupPreview>('InspectBackup', path),
  restoreBackup: (path: string) => call<RestoreResult>('RestoreBackup', path),
  listAutomaticBackups: () => call<string[]>('ListAutomaticBackups'),
  runIntegrityCheck: () => call<IntegrityReport>('RunIntegrityCheck'),

  hasDemoData: () => call<boolean>('HasDemoData'),
  loadDemoData: () => call<void>('LoadDemoData'),
  removeDemoData: () => call<void>('RemoveDemoData'),
}

export function errorMessage(err: unknown): string {
  if (err instanceof Error) return err.message
  if (typeof err === 'string') return err
  return 'Ocorreu um erro inesperado.'
}

export function fieldErrors(err: unknown): Record<string, string> {
  return err instanceof ApiError ? err.fields : {}
}
