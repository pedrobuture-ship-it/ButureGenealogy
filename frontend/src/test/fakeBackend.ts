import type { Animal, AnimalInput, Farm, FarmInput } from '../types/domain'

type Fn = (...a: never[]) => Promise<unknown>

/** Minimal in-memory stand-in for the Wails Go bindings, for UI tests. */
export function installFakeBackend(seed: { farms?: Farm[]; animals?: Animal[]; traitGroups?: unknown[]; libraries?: Record<string, unknown>[]; inspection?: unknown; validation?: unknown; lots?: Record<string, unknown>[]; profiles?: unknown[]; runResponse?: unknown; calculations?: Record<string, unknown>[]; inspectBackup?: unknown; integrity?: unknown; demo?: boolean } = {}) {
  const traitGroups = seed.traitGroups ?? []
  const libraries = [...(seed.libraries ?? [])]
  const profiles = [...(seed.profiles ?? [])]
  let demo = seed.demo ?? false
  const lots = (seed.lots ?? []) as { id: string; bullId: string; supplier: string; batch: string; quantity: number; unitCost: number | null; semenType: string; notes: string }[]
  const movements: Record<string, unknown>[] = []
  const stockStatus = (bullId: string) => {
    const ls = lots.filter((l) => l.bullId === bullId)
    return ls.length === 0 ? 'no_semen' : ls.reduce((t, l) => t + l.quantity, 0) === 0 ? 'depleted' : 'in_stock'
  }
  const moveLot = (kind: string, in_: { inventoryId: string; quantity: number; reason: string; note: string }) => {
    const lot = lots.find((l) => l.id === in_.inventoryId)!
    const delta = kind === 'withdrawal' ? -in_.quantity : in_.quantity
    if (lot.quantity + delta < 0) return wireErr('Verifique os campos destacados.', { quantity: `Estoque insuficiente: há ${lot.quantity} dose(s) neste lote.` })
    lot.quantity += delta
    const m = { id: `m${movements.length + 1}`, inventoryId: lot.id, bullId: lot.bullId, type: kind, delta, balanceAfter: lot.quantity, reason: in_.reason, note: in_.note, matingId: '', isDemo: false, createdAt: '2026-10-11T12:00:00Z' }
    movements.unshift(m)
    return m
  }
  const farms: Farm[] = [...(seed.farms ?? [])]
  const animals: Animal[] = [...(seed.animals ?? [])]
  const calls: { method: string; args: unknown[] }[] = []
  let n = 100

  const wireErr = (message: string, fields: Record<string, string> = {}) =>
    Promise.reject(JSON.stringify({ code: 'validation', message, fields }))

  const impl: Record<string, (...a: any[]) => unknown> = { // eslint-disable-line @typescript-eslint/no-explicit-any
    GetAppInfo: () => ({ name: 'Genetica', version: '0.1.0', dataDir: '/tmp', platform: 'linux' }),
    HealthCheck: () => ({ ok: true, path: '/tmp/farm.db', schemaVersion: 2, journalMode: 'wal', foreignKeys: true, quickCheck: 'ok' }),
    ListFarms: () => farms,
    CreateFarm: (in_: FarmInput) => {
      if (!in_.name.trim()) return wireErr('Verifique os campos destacados.', { name: 'Informe o nome da propriedade.' })
      const f: Farm = { id: `f${++n}`, ...in_, isDemo: false, createdAt: '', updatedAt: '' }
      farms.push(f)
      return f
    },
    ListBreeds: () => [{ id: 'breed-nelore', code: 'NEL', name: 'Nelore', species: 'Bos indicus', purpose: 'beef', active: true }],
    ListAnimals: (q: { search?: string; sex?: string; includeArchived?: boolean; status?: string }) => {
      let items = animals.filter((a) => (q.includeArchived || q.status ? true : a.status !== 'archived'))
      if (q.sex) items = items.filter((a) => a.sex === q.sex)
      if (q.search) items = items.filter((a) => a.name.toLowerCase().includes(q.search!.toLowerCase()))
      return { items, total: items.length, page: 1 }
    },
    GetAnimal: (id: string) => animals.find((x) => x.id === id) ?? Promise.reject(JSON.stringify({ code: 'not_found', message: 'Animal não encontrado.' })),
    GetAnimalDetail: (id: string) => {
      const a = animals.find((x) => x.id === id)
      if (!a) return Promise.reject(JSON.stringify({ code: 'not_found', message: 'Animal não encontrado.' }))
      return { animal: a, breedName: 'Nelore', farmName: farms[0]?.name ?? '', father: null, mother: null, offspring: [], semenDoses: 0 }
    },
    CreateAnimal: (in_: AnimalInput) => {
      if (!in_.name && !in_.registration) return wireErr('Verifique os campos destacados.', { name: 'Informe o nome ou o registro do animal.' })
      const a: Animal = { id: `a${++n}`, ...in_, isDemo: false, createdAt: '', updatedAt: '' }
      animals.push(a)
      return a
    },
    UpdateAnimal: (id: string, in_: AnimalInput) => {
      const a = animals.find((x) => x.id === id)!
      Object.assign(a, in_)
      return a
    },
    ArchiveAnimal: (id: string) => { animals.find((x) => x.id === id)!.status = 'archived' },
    RestoreAnimal: (id: string) => { animals.find((x) => x.id === id)!.status = 'active' },
    GetPedigree: (id: string, generations: number) => {
      const node = (aid: string, gen: number, role: 'self' | 'sire' | 'dam', childId: string): unknown => {
        const a = animals.find((x) => x.id === aid)
        if (!a) return { id: '', name: '', registration: '', sex: '', breedName: '', birthDate: '', generation: gen, role, known: false, childId, childName: animals.find((x) => x.id === childId)?.name ?? '', hasMore: false, sire: null, dam: null }
        const more = gen < generations
        return { id: a.id, name: a.name, registration: a.registration, sex: a.sex, breedName: 'Nelore', birthDate: a.birthDate, generation: gen, role, known: true, childId, hasMore: false,
          sire: more ? node(a.fatherId, gen + 1, 'sire', a.id) : null, dam: more ? node(a.motherId, gen + 1, 'dam', a.id) : null }
      }
      return { animalId: id, generations, tree: node(id, 0, 'self', ''), depth: 1, repeated: [], issues: [],
        stats: [{ generation: 1, expected: 2, known: 1, percent: 50 }],
        inbreeding: { status: 'not_calculated', value: null, method: '', message: 'não calculado nesta versão' } }
    },
    CreateMinimalAncestor: (in_: { childId: string; role: 'sire' | 'dam'; name: string; registration: string }) => {
      if (!in_.name && !in_.registration) return wireErr('Verifique os campos destacados.', { name: 'Informe o nome ou o registro do animal.' })
      const child = animals.find((x) => x.id === in_.childId)!
      const a: Animal = { id: `a${++n}`, farmId: child.farmId, registration: in_.registration, name: in_.name, sex: in_.role === 'sire' ? 'male' : 'female', breedId: '', birthDate: '',
        category: in_.role === 'sire' ? 'bull' : 'cow', status: 'active', fatherId: '', motherId: '', notes: '', isDemo: false, createdAt: '', updatedAt: '' }
      animals.push(a)
      if (in_.role === 'sire') child.fatherId = a.id; else child.motherId = a.id
      return a
    },
    ListTraits: () => [{ id: 't1', code: 'BODY_WEIGHT', name: 'Peso corporal', unit: 'kg', evaluationType: 'PHENOTYPE', category: 'growth' }],
    GetAnimalTraits: () => traitGroups,
    PickLibraryFile: () => '/tmp/avaliacoes.csv',
    ListLibraries: () => libraries,
    GetLibrary: (id: string) => ({ ...libraries.find((l) => l.id === id), manifest: {}, readme: '# Documentação de teste', licenseText: 'Texto da licença de teste', traits: [], usedByCalculations: 0,
      responsibilityNotice: 'A responsabilidade pela autorização de uso e distribuição de dados importados manualmente é do usuário.' }),
    SetLibraryActive: (id: string, active: boolean) => { libraries.find((l) => l.id === id)!.active = active },
    RemoveLibrary: (id: string) => { libraries.splice(libraries.findIndex((l) => l.id === id), 1) },
    InspectImport: () => seed.inspection,
    ValidateImport: () => seed.validation,
    InstallLibrary: (req: { manifest: { name: string; version: string } }) => {
      if (!(req as unknown as { acceptResponsibility: boolean }).acceptResponsibility) return wireErr('Confirme', { acceptResponsibility: 'Confirme para continuar.' })
      libraries.push({ id: 'new', name: req.manifest.name, version: req.manifest.version, packageId: 'x', origin: 'csv', breeds: [], purposes: [], organization: '', licenseStatus: 'unknown', licenseName: '', active: true, recordCount: 3, traitCount: 1 })
      return { library: libraries[libraries.length - 1], report: {}, evaluationsInserted: 6, animalsCreated: 0, recordsSkipped: 1 }
    },
    ListValueKinds: () => [{ kind: 'DEP', label: 'DEP', description: 'desc', genetic: true }],
    SemenBank: (q: { stockFilter?: string; search?: string }) => {
      let bulls = animals.filter((a) => a.sex === 'male' && a.status !== 'archived')
      const filter = q.stockFilter || 'with_lots'
      bulls = bulls.filter((a) => (filter === 'all' ? true : filter === 'in_stock' ? stockStatus(a.id) === 'in_stock' : lots.some((l) => l.bullId === a.id)))
      if (q.search) bulls = bulls.filter((a) => a.name.toLowerCase().includes(q.search!.toLowerCase()))
      const items = bulls.map((a) => {
        const ls = lots.filter((l) => l.bullId === a.id)
        const doses = ls.reduce((t, l) => t + l.quantity, 0)
        return { id: a.id, name: a.name, registration: a.registration, breedName: 'Nelore', animalStatus: a.status, doses, lots: ls.length,
          suppliers: [...new Set(ls.map((l) => l.supplier).filter(Boolean))], unitCost: ls.find((l) => l.unitCost != null)?.unitCost ?? null, stockStatus: stockStatus(a.id), isDemo: false }
      })
      return { items, total: items.length, page: 1, totalDoses: items.reduce((t, r) => t + r.doses, 0), suppliers: [...new Set(lots.map((l) => l.supplier))] }
    },
    GetBullSheet: (id: string) => {
      const a = animals.find((x) => x.id === id)
      if (!a) return Promise.reject(JSON.stringify({ code: 'not_found', message: 'Animal não encontrado.' }))
      const ls = lots.filter((l) => l.bullId === id)
      return { detail: { animal: a, breedName: 'Nelore', farmName: farms[0]?.name ?? '', father: null, mother: null, offspring: [], semenDoses: 0 }, lots: ls,
        totalDoses: ls.reduce((t, l) => t + l.quantity, 0), unitCost: ls.find((l) => l.unitCost != null)?.unitCost ?? null, movements: movements.filter((m) => m.bullId === id), stockStatus: stockStatus(id) }
    },
    ListSemenTypes: () => [{ code: 'conventional', name: 'Convencional', isBuiltin: true }, { code: 'sexed', name: 'Sexado', isBuiltin: true }],
    CreateSemenType: (name: string) => ({ code: name.toLowerCase().replace(/\s+/g, '-'), name, isBuiltin: false }),
    AddSemenLot: (in_: { bullId: string; supplier: string; batch: string; quantity: number; unitCost: number | null; semenType: string; notes: string }) => {
      const lot = { id: `l${lots.length + 1}`, bullId: in_.bullId, supplier: in_.supplier, batch: in_.batch, quantity: in_.quantity, unitCost: in_.unitCost, semenType: in_.semenType, notes: in_.notes }
      lots.push(lot)
      return lot
    },
    UpdateSemenLot: (id: string, in_: Record<string, unknown>) => Object.assign(lots.find((l) => l.id === id)!, in_),
    ReceiveSemen: (in_: never) => moveLot('entry', in_),
    WithdrawSemen: (in_: never) => moveLot('withdrawal', in_),
    AdjustSemen: (in_: never) => moveLot('adjustment', in_),
    ListMatingProfiles: () => profiles,
    SaveMatingProfile: (p: { id: string; name: string }) => { const saved = { ...p, id: p.id || `saved-${profiles.length}` }; profiles.push(saved); return saved },
    DeleteMatingProfile: (id: string) => { profiles.splice(profiles.findIndex((x) => (x as { id: string }).id === id), 1) },
    RunMating: () => seed.runResponse,
    GetCalculation: (id: string) => (seed.calculations ?? []).find((c) => c.id === id),
    ListCalculations: () => (seed.calculations ?? []).map((c) => ({ id: c.id, runId: 'run1', bullId: 'b', bullName: 'Touro A', profileName: 'Personalizada', status: 'draft', score: 65.8, rank: 1, engineVersion: '0.1.0', createdAt: '2026-10-11T12:00:00Z', isDemo: false })),
    EngineVersion: () => '0.1.0',
    SuggestedBackupName: () => '2026-10-11-fazenda.genbackup',
    PickBackupSavePath: (name: string) => `/home/u/${name}`,
    PickBackupOpenPath: () => '/home/u/antigo.genbackup',
    CreateBackup: () => ({ formatVersion: 1, createdAt: '2026-10-11T12:00:00Z', appVersion: '0.1.0', engineVersion: '0.1.0', schemaVersion: 6, farms: ['Fazenda'], counts: { animals: 12 }, files: [], libraries: [{ id: 'l' }] }),
    InspectBackup: () => seed.inspectBackup,
    RestoreBackup: () => ({ safetyBackup: '/data/backups/pre-restore-20261011-120000.genbackup', counts: {}, missingLibraries: [] }),
    ListAutomaticBackups: () => [],
    RunIntegrityCheck: () => seed.integrity,
    HasDemoData: () => demo,
    LoadDemoData: () => { demo = true },
    RemoveDemoData: () => { demo = false },
    ListMeasurements: () => [],
    ListEvaluations: () => [],
  }

  const App: Record<string, Fn> = {}
  for (const [name, fn] of Object.entries(impl)) {
    App[name] = (async (...args: unknown[]) => {
      calls.push({ method: name, args })
      return fn(...args)
    }) as unknown as Fn
  }
  window.go = { main: { App: App as unknown as Record<string, (...a: unknown[]) => Promise<unknown>> } }
  return { farms, animals, calls, lots, movements }
}

export const sampleFarm: Farm = { id: 'f1', name: 'Fazenda Teste', ownerName: '', city: '', state: '', notes: '', isDemo: false, createdAt: '', updatedAt: '' }

export function sampleAnimal(over: Partial<Animal> = {}): Animal {
  return {
    id: 'a1', farmId: 'f1', registration: 'R-1', name: 'Mimosa', sex: 'female', breedId: 'breed-nelore', birthDate: '2020-01-15',
    category: 'cow', status: 'active', fatherId: '', motherId: '', notes: '', isDemo: false, createdAt: '', updatedAt: '', ...over,
  }
}
