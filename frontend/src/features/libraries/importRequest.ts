import type { ImportRequest, LibraryManifest, TraitDef } from '../../types/domain'

export interface MetaForm {
  name: string
  id: string
  version: string
  species: string
  breeds: string
  purposes: string[]
  organization: string
  sourceUrl: string
  releaseDate: string
  program: string
  release: string
  geneticBase: string
  licenseName: string
  licenseUrl: string
  redistributionAllowed: boolean
  licenseText: string
  idColumn: string
  nameColumn: string
  sexColumn: string
  defaultSex: 'male' | 'female'
  decimalComma: boolean
  accuracyScale: 'fraction' | 'percent'
}

export const emptyMeta: MetaForm = {
  name: '', id: '', version: '', species: 'Bos taurus', breeds: '', purposes: ['beef'], organization: '', sourceUrl: '',
  releaseDate: '', program: '', release: '', geneticBase: '', licenseName: '', licenseUrl: '', redistributionAllowed: false,
  licenseText: '', idColumn: '', nameColumn: '', sexColumn: '', defaultSex: 'male', decimalComma: false, accuracyScale: 'fraction',
}

/** One data column the user maps to a trait. */
export interface ColumnMapping {
  column: string
  enabled: boolean
  code: string
  name: string
  evaluationType: string
  unit: string
  category: string
  accuracyColumn: string
  percentileColumn: string
  existingTraitId: string
  /** What the percentile means; the engine refuses to use a percentile without it. */
  percentileSemantics: '' | 'ascending' | 'descending'
  popMean: string
  popSd: string
}

export function slugify(s: string): string {
  return s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '').slice(0, 60)
}

/** Suggests a trait code from a column name, e.g. "DEP_PD" -> "PD". Pure convenience. */
export function suggestCode(column: string): string {
  const c = column.replace(/[^A-Za-z0-9_]/g, '_').replace(/^_+/, '')
  return /^[A-Za-z]/.test(c) ? c : `C_${c}`
}

export function buildManifest(m: MetaForm): LibraryManifest {
  const id = m.id || slugify(m.name)
  return {
    schemaVersion: 1,
    id,
    name: m.name.trim(),
    version: m.version.trim(),
    species: m.species.trim(),
    breeds: m.breeds.split(',').map((b) => b.trim()).filter(Boolean),
    purpose: m.purposes,
    authors: [],
    organization: m.organization.trim(),
    sourceUrl: m.sourceUrl.trim(),
    releaseDate: m.releaseDate,
    // Nothing is assumed: an empty license name stays "unknown" and redistribution stays false unless ticked.
    license: { name: m.licenseName.trim(), url: m.licenseUrl.trim(), redistributionAllowed: m.licenseName.trim() !== '' && m.redistributionAllowed, commercialUseAllowed: null },
    evaluation: { program: m.program.trim(), release: m.release.trim(), geneticBase: m.geneticBase.trim() },
    data: {
      format: '', idColumn: m.idColumn, nameColumn: m.nameColumn, sexColumn: m.sexColumn, breedColumn: '',
      defaultSex: m.sexColumn ? '' : m.defaultSex, decimalComma: m.decimalComma, accuracyScale: m.accuracyScale,
    },
  }
}

const num = (t: string): number | null => {
  const v = t.trim().replace(',', '.')
  if (v === '') return null
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

/** Trait metadata the engine understands: percentile semantics and population parameters declared by the user. */
export function buildMetadata(x: Pick<ColumnMapping, 'percentileColumn' | 'percentileSemantics' | 'popMean' | 'popSd'>): Record<string, unknown> {
  const meta: Record<string, unknown> = {}
  if (x.percentileColumn && x.percentileSemantics) meta.percentileSemantics = x.percentileSemantics
  const mean = num(x.popMean)
  const sd = num(x.popSd)
  if (mean !== null && sd !== null && sd > 0) meta.population = { mean, sd }
  return meta
}

export function buildTraits(mappings: ColumnMapping[]): TraitDef[] {
  return mappings.filter((x) => x.enabled).map((x) => ({
    code: x.code.trim(), name: x.name.trim(), description: '', category: x.category, purpose: 'any', unit: x.unit.trim(), dataType: 'number',
    evaluationType: x.evaluationType, optimizationDirection: 'INFORMATION_ONLY', idealMin: null, idealMax: null, higherIsBetter: null, metadata: buildMetadata(x),
    column: x.column, accuracyColumn: x.accuracyColumn, percentileColumn: x.percentileColumn,
    ...(x.existingTraitId ? { existingTraitId: x.existingTraitId } : {}),
  }))
}

export function buildRequest(path: string, meta: MetaForm, mappings: ColumnMapping[]): ImportRequest {
  return {
    path, manifest: buildManifest(meta), traits: buildTraits(mappings), licenseText: meta.licenseText,
    unmatchedPolicy: 'skip', farmId: '', acceptResponsibility: false,
  }
}
