export namespace backup {
	
	export class FileEntry {
	    path: string;
	    size: number;
	    sha256: string;
	
	    static createFrom(source: any = {}) {
	        return new FileEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.size = source["size"];
	        this.sha256 = source["sha256"];
	    }
	}
	export class Issue {
	    severity: string;
	    code: string;
	    message: string;
	    detail: string;
	
	    static createFrom(source: any = {}) {
	        return new Issue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.severity = source["severity"];
	        this.code = source["code"];
	        this.message = source["message"];
	        this.detail = source["detail"];
	    }
	}
	export class LibraryEntry {
	    id: string;
	    packageId: string;
	    name: string;
	    version: string;
	    fileName: string;
	    sha256: string;
	    licenseStatus: string;
	    isDemo: boolean;
	    isOfficial: boolean;
	    included: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LibraryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.packageId = source["packageId"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.fileName = source["fileName"];
	        this.sha256 = source["sha256"];
	        this.licenseStatus = source["licenseStatus"];
	        this.isDemo = source["isDemo"];
	        this.isOfficial = source["isOfficial"];
	        this.included = source["included"];
	    }
	}
	export class Manifest {
	    formatVersion: number;
	    createdAt: string;
	    appVersion: string;
	    engineVersion: string;
	    schemaVersion: number;
	    farms: string[];
	    counts: Record<string, number>;
	    files: FileEntry[];
	    libraries: LibraryEntry[];
	
	    static createFrom(source: any = {}) {
	        return new Manifest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.formatVersion = source["formatVersion"];
	        this.createdAt = source["createdAt"];
	        this.appVersion = source["appVersion"];
	        this.engineVersion = source["engineVersion"];
	        this.schemaVersion = source["schemaVersion"];
	        this.farms = source["farms"];
	        this.counts = source["counts"];
	        this.files = this.convertValues(source["files"], FileEntry);
	        this.libraries = this.convertValues(source["libraries"], LibraryEntry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Preview {
	    manifest: Manifest;
	    upgradedFrom: number;
	    schemaVersion: number;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new Preview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.manifest = this.convertValues(source["manifest"], Manifest);
	        this.upgradedFrom = source["upgradedFrom"];
	        this.schemaVersion = source["schemaVersion"];
	        this.path = source["path"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Report {
	    ok: boolean;
	    checkedAt: string;
	    errors: number;
	    warnings: number;
	    issues: Issue[];
	    checked: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new Report(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.checkedAt = source["checkedAt"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	        this.issues = this.convertValues(source["issues"], Issue);
	        this.checked = source["checked"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RestoreResult {
	    safetyBackup: string;
	    counts: Record<string, number>;
	    missingLibraries: string[];
	
	    static createFrom(source: any = {}) {
	        return new RestoreResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.safetyBackup = source["safetyBackup"];
	        this.counts = source["counts"];
	        this.missingLibraries = source["missingLibraries"];
	    }
	}

}

export namespace database {
	
	export class Health {
	    ok: boolean;
	    path: string;
	    schemaVersion: number;
	    journalMode: string;
	    foreignKeys: boolean;
	    quickCheck: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new Health(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.path = source["path"];
	        this.schemaVersion = source["schemaVersion"];
	        this.journalMode = source["journalMode"];
	        this.foreignKeys = source["foreignKeys"];
	        this.quickCheck = source["quickCheck"];
	        this.error = source["error"];
	    }
	}

}

export namespace datasets {
	
	export class DataSpec {
	    format: string;
	    idColumn: string;
	    nameColumn: string;
	    sexColumn: string;
	    breedColumn: string;
	    defaultSex: string;
	    decimalComma: boolean;
	    accuracyScale: string;
	
	    static createFrom(source: any = {}) {
	        return new DataSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.idColumn = source["idColumn"];
	        this.nameColumn = source["nameColumn"];
	        this.sexColumn = source["sexColumn"];
	        this.breedColumn = source["breedColumn"];
	        this.defaultSex = source["defaultSex"];
	        this.decimalComma = source["decimalComma"];
	        this.accuracyScale = source["accuracyScale"];
	    }
	}
	export class Evaluation {
	    program: string;
	    release: string;
	    geneticBase: string;
	
	    static createFrom(source: any = {}) {
	        return new Evaluation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.program = source["program"];
	        this.release = source["release"];
	        this.geneticBase = source["geneticBase"];
	    }
	}
	export class License {
	    name: string;
	    url: string;
	    redistributionAllowed: boolean;
	    commercialUseAllowed?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new License(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.redistributionAllowed = source["redistributionAllowed"];
	        this.commercialUseAllowed = source["commercialUseAllowed"];
	    }
	}
	export class Manifest {
	    schemaVersion: number;
	    id: string;
	    name: string;
	    version: string;
	    species: string;
	    breeds: string[];
	    purpose: string[];
	    authors: string[];
	    organization: string;
	    sourceUrl: string;
	    releaseDate: string;
	    license: License;
	    evaluation: Evaluation;
	    data: DataSpec;
	
	    static createFrom(source: any = {}) {
	        return new Manifest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemaVersion = source["schemaVersion"];
	        this.id = source["id"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.species = source["species"];
	        this.breeds = source["breeds"];
	        this.purpose = source["purpose"];
	        this.authors = source["authors"];
	        this.organization = source["organization"];
	        this.sourceUrl = source["sourceUrl"];
	        this.releaseDate = source["releaseDate"];
	        this.license = this.convertValues(source["license"], License);
	        this.evaluation = this.convertValues(source["evaluation"], Evaluation);
	        this.data = this.convertValues(source["data"], DataSpec);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Report {
	    records: number;
	    rows: number;
	    rowsWithoutId: number;
	    duplicates: number;
	    values: number;
	    traitsRecognized: number;
	    traitsUnrecognized: number;
	    unrecognizedColumns: string[];
	    invalidValues: number;
	    withoutAccuracy: number;
	    accuracyOutOfRange: number;
	    percentileInvalid: number;
	    matchedAnimals: number;
	    unmatchedAnimals: number;
	    ambiguousAnimals: number;
	    errors: string[];
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new Report(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.records = source["records"];
	        this.rows = source["rows"];
	        this.rowsWithoutId = source["rowsWithoutId"];
	        this.duplicates = source["duplicates"];
	        this.values = source["values"];
	        this.traitsRecognized = source["traitsRecognized"];
	        this.traitsUnrecognized = source["traitsUnrecognized"];
	        this.unrecognizedColumns = source["unrecognizedColumns"];
	        this.invalidValues = source["invalidValues"];
	        this.withoutAccuracy = source["withoutAccuracy"];
	        this.accuracyOutOfRange = source["accuracyOutOfRange"];
	        this.percentileInvalid = source["percentileInvalid"];
	        this.matchedAnimals = source["matchedAnimals"];
	        this.unmatchedAnimals = source["unmatchedAnimals"];
	        this.ambiguousAnimals = source["ambiguousAnimals"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	    }
	}
	export class TraitDef {
	    code: string;
	    name: string;
	    description: string;
	    category: string;
	    purpose: string;
	    unit: string;
	    dataType: string;
	    evaluationType: string;
	    optimizationDirection: string;
	    idealMin?: number;
	    idealMax?: number;
	    higherIsBetter?: boolean;
	    metadata: Record<string, any>;
	    column: string;
	    accuracyColumn: string;
	    percentileColumn: string;
	    existingTraitId?: string;
	
	    static createFrom(source: any = {}) {
	        return new TraitDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.purpose = source["purpose"];
	        this.unit = source["unit"];
	        this.dataType = source["dataType"];
	        this.evaluationType = source["evaluationType"];
	        this.optimizationDirection = source["optimizationDirection"];
	        this.idealMin = source["idealMin"];
	        this.idealMax = source["idealMax"];
	        this.higherIsBetter = source["higherIsBetter"];
	        this.metadata = source["metadata"];
	        this.column = source["column"];
	        this.accuracyColumn = source["accuracyColumn"];
	        this.percentileColumn = source["percentileColumn"];
	        this.existingTraitId = source["existingTraitId"];
	    }
	}

}

export namespace domain {
	
	export class Animal {
	    id: string;
	    farmId: string;
	    registration: string;
	    name: string;
	    sex: string;
	    breedId: string;
	    birthDate: string;
	    category: string;
	    status: string;
	    fatherId: string;
	    motherId: string;
	    notes: string;
	    isDemo: boolean;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Animal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.farmId = source["farmId"];
	        this.registration = source["registration"];
	        this.name = source["name"];
	        this.sex = source["sex"];
	        this.breedId = source["breedId"];
	        this.birthDate = source["birthDate"];
	        this.category = source["category"];
	        this.status = source["status"];
	        this.fatherId = source["fatherId"];
	        this.motherId = source["motherId"];
	        this.notes = source["notes"];
	        this.isDemo = source["isDemo"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class Breed {
	    id: string;
	    code: string;
	    name: string;
	    species: string;
	    purpose: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Breed(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.species = source["species"];
	        this.purpose = source["purpose"];
	        this.active = source["active"];
	    }
	}
	export class Farm {
	    id: string;
	    name: string;
	    ownerName: string;
	    city: string;
	    state: string;
	    notes: string;
	    isDemo: boolean;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Farm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.ownerName = source["ownerName"];
	        this.city = source["city"];
	        this.state = source["state"];
	        this.notes = source["notes"];
	        this.isDemo = source["isDemo"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class GeneticEvaluation {
	    id: string;
	    animalId: string;
	    datasetId: string;
	    evaluationProgram: string;
	    evaluationRelease: string;
	    breedCode: string;
	    traitCode: string;
	    evaluationType: string;
	    value: number;
	    unit: string;
	    accuracy?: number;
	    percentile?: number;
	    geneticBase: string;
	    sourceReference: string;
	    evaluatedAt: string;
	    isDemo: boolean;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new GeneticEvaluation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.animalId = source["animalId"];
	        this.datasetId = source["datasetId"];
	        this.evaluationProgram = source["evaluationProgram"];
	        this.evaluationRelease = source["evaluationRelease"];
	        this.breedCode = source["breedCode"];
	        this.traitCode = source["traitCode"];
	        this.evaluationType = source["evaluationType"];
	        this.value = source["value"];
	        this.unit = source["unit"];
	        this.accuracy = source["accuracy"];
	        this.percentile = source["percentile"];
	        this.geneticBase = source["geneticBase"];
	        this.sourceReference = source["sourceReference"];
	        this.evaluatedAt = source["evaluatedAt"];
	        this.isDemo = source["isDemo"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class Measurement {
	    id: string;
	    animalId: string;
	    traitCode: string;
	    value: number;
	    unit: string;
	    measuredAt: string;
	    source: string;
	    notes: string;
	    isDemo: boolean;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Measurement(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.animalId = source["animalId"];
	        this.traitCode = source["traitCode"];
	        this.value = source["value"];
	        this.unit = source["unit"];
	        this.measuredAt = source["measuredAt"];
	        this.source = source["source"];
	        this.notes = source["notes"];
	        this.isDemo = source["isDemo"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class SemenInventory {
	    id: string;
	    bullId: string;
	    supplier: string;
	    batch: string;
	    quantity: number;
	    unitCost?: number;
	    semenType: string;
	    notes: string;
	    isDemo: boolean;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SemenInventory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.bullId = source["bullId"];
	        this.supplier = source["supplier"];
	        this.batch = source["batch"];
	        this.quantity = source["quantity"];
	        this.unitCost = source["unitCost"];
	        this.semenType = source["semenType"];
	        this.notes = source["notes"];
	        this.isDemo = source["isDemo"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class SemenMovement {
	    id: string;
	    inventoryId: string;
	    bullId: string;
	    type: string;
	    delta: number;
	    balanceAfter: number;
	    reason: string;
	    note: string;
	    matingId: string;
	    isDemo: boolean;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SemenMovement(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.inventoryId = source["inventoryId"];
	        this.bullId = source["bullId"];
	        this.type = source["type"];
	        this.delta = source["delta"];
	        this.balanceAfter = source["balanceAfter"];
	        this.reason = source["reason"];
	        this.note = source["note"];
	        this.matingId = source["matingId"];
	        this.isDemo = source["isDemo"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class SemenType {
	    code: string;
	    name: string;
	    isBuiltin: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SemenType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.isBuiltin = source["isBuiltin"];
	    }
	}
	export class Trait {
	    id: string;
	    code: string;
	    name: string;
	    description: string;
	    category: string;
	    purpose: string;
	    unit: string;
	    dataType: string;
	    evaluationType: string;
	    optimizationDirection: string;
	    idealMin?: number;
	    idealMax?: number;
	    higherIsBetter?: boolean;
	    sourceDataset: string;
	    metadata: string;
	    isBuiltin: boolean;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Trait(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.code = source["code"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.purpose = source["purpose"];
	        this.unit = source["unit"];
	        this.dataType = source["dataType"];
	        this.evaluationType = source["evaluationType"];
	        this.optimizationDirection = source["optimizationDirection"];
	        this.idealMin = source["idealMin"];
	        this.idealMax = source["idealMax"];
	        this.higherIsBetter = source["higherIsBetter"];
	        this.sourceDataset = source["sourceDataset"];
	        this.metadata = source["metadata"];
	        this.isBuiltin = source["isBuiltin"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class ValueKindInfo {
	    kind: string;
	    label: string;
	    description: string;
	    genetic: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ValueKindInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.genetic = source["genetic"];
	    }
	}

}

export namespace engine {
	
	export class Animal {
	    id: string;
	    name: string;
	    registration: string;
	    sex: string;
	    status: string;
	    breedName: string;
	    isDemo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Animal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.registration = source["registration"];
	        this.sex = source["sex"];
	        this.status = source["status"];
	        this.breedName = source["breedName"];
	        this.isDemo = source["isDemo"];
	    }
	}
	export class Basis {
	    traitCode: string;
	    evaluationType: string;
	    program: string;
	    release: string;
	    geneticBase: string;
	    breedCode: string;
	    unit: string;
	
	    static createFrom(source: any = {}) {
	        return new Basis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.traitCode = source["traitCode"];
	        this.evaluationType = source["evaluationType"];
	        this.program = source["program"];
	        this.release = source["release"];
	        this.geneticBase = source["geneticBase"];
	        this.breedCode = source["breedCode"];
	        this.unit = source["unit"];
	    }
	}
	export class BasisCount {
	    basis: Basis;
	    candidates: number;
	
	    static createFrom(source: any = {}) {
	        return new BasisCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.basis = this.convertValues(source["basis"], Basis);
	        this.candidates = source["candidates"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FinalScore {
	    base: number;
	    complementarity: number;
	    penalties: number;
	    unclamped: number;
	    final: number;
	    clamped: boolean;
	    coverage: number;
	    formula: string;
	
	    static createFrom(source: any = {}) {
	        return new FinalScore(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.base = source["base"];
	        this.complementarity = source["complementarity"];
	        this.penalties = source["penalties"];
	        this.unclamped = source["unclamped"];
	        this.final = source["final"];
	        this.clamped = source["clamped"];
	        this.coverage = source["coverage"];
	        this.formula = source["formula"];
	    }
	}
	export class PenaltyResult {
	    name: string;
	    status: string;
	    points: number;
	    formula: string;
	    source: Source;
	    explanation: string;
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new PenaltyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.points = source["points"];
	        this.formula = source["formula"];
	        this.source = this.convertValues(source["source"], Source);
	        this.explanation = source["explanation"];
	        this.warnings = source["warnings"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ComplementarityItem {
	    traitCode: string;
	    name: string;
	    status: string;
	    bullScore?: number;
	    damScore?: number;
	    gain?: number;
	    weight: number;
	    points: number;
	    explanation: string;
	    source: Source;
	
	    static createFrom(source: any = {}) {
	        return new ComplementarityItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.traitCode = source["traitCode"];
	        this.name = source["name"];
	        this.status = source["status"];
	        this.bullScore = source["bullScore"];
	        this.damScore = source["damScore"];
	        this.gain = source["gain"];
	        this.weight = source["weight"];
	        this.points = source["points"];
	        this.explanation = source["explanation"];
	        this.source = this.convertValues(source["source"], Source);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ComplementarityResult {
	    status: string;
	    points: number;
	    maxBonus: number;
	    formula: string;
	    items: ComplementarityItem[];
	    source: Source;
	    explanation: string;
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new ComplementarityResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.points = source["points"];
	        this.maxBonus = source["maxBonus"];
	        this.formula = source["formula"];
	        this.items = this.convertValues(source["items"], ComplementarityItem);
	        this.source = this.convertValues(source["source"], Source);
	        this.explanation = source["explanation"];
	        this.warnings = source["warnings"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Source {
	    evaluationId: string;
	    datasetId: string;
	    datasetName: string;
	    datasetVersion: string;
	    program: string;
	    release: string;
	    geneticBase: string;
	    isDemo: boolean;
	    parameter: string;
	
	    static createFrom(source: any = {}) {
	        return new Source(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.evaluationId = source["evaluationId"];
	        this.datasetId = source["datasetId"];
	        this.datasetName = source["datasetName"];
	        this.datasetVersion = source["datasetVersion"];
	        this.program = source["program"];
	        this.release = source["release"];
	        this.geneticBase = source["geneticBase"];
	        this.isDemo = source["isDemo"];
	        this.parameter = source["parameter"];
	    }
	}
	export class ComponentResult {
	    name: string;
	    traitCode: string;
	    evaluationType: string;
	    status: string;
	    direction: string;
	    inputValue?: number;
	    damValue?: number;
	    unit: string;
	    accuracy?: number;
	    percentile?: number;
	    score?: number;
	    method: string;
	    transformation: string;
	    formula: string;
	    rawWeight: number;
	    weight: number;
	    contribution?: number;
	    basis?: Basis;
	    source: Source;
	    explanation: string;
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new ComponentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.traitCode = source["traitCode"];
	        this.evaluationType = source["evaluationType"];
	        this.status = source["status"];
	        this.direction = source["direction"];
	        this.inputValue = source["inputValue"];
	        this.damValue = source["damValue"];
	        this.unit = source["unit"];
	        this.accuracy = source["accuracy"];
	        this.percentile = source["percentile"];
	        this.score = source["score"];
	        this.method = source["method"];
	        this.transformation = source["transformation"];
	        this.formula = source["formula"];
	        this.rawWeight = source["rawWeight"];
	        this.weight = source["weight"];
	        this.contribution = source["contribution"];
	        this.basis = this.convertValues(source["basis"], Basis);
	        this.source = this.convertValues(source["source"], Source);
	        this.explanation = source["explanation"];
	        this.warnings = source["warnings"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Exclusion {
	    filter: string;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new Exclusion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filter = source["filter"];
	        this.reason = source["reason"];
	    }
	}
	export class Stock {
	    doses: number;
	    unitCost?: number;
	
	    static createFrom(source: any = {}) {
	        return new Stock(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.doses = source["doses"];
	        this.unitCost = source["unitCost"];
	    }
	}
	export class CandidateResult {
	    bull: Animal;
	    stock: Stock;
	    status: string;
	    rank: number;
	    tied: boolean;
	    exclusions: Exclusion[];
	    reasons: string[];
	    components: ComponentResult[];
	    complementarity: ComplementarityResult;
	    penalties: PenaltyResult[];
	    final?: FinalScore;
	    warnings: string[];
	    pedigreeNotes: string[];
	
	    static createFrom(source: any = {}) {
	        return new CandidateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bull = this.convertValues(source["bull"], Animal);
	        this.stock = this.convertValues(source["stock"], Stock);
	        this.status = source["status"];
	        this.rank = source["rank"];
	        this.tied = source["tied"];
	        this.exclusions = this.convertValues(source["exclusions"], Exclusion);
	        this.reasons = source["reasons"];
	        this.components = this.convertValues(source["components"], ComponentResult);
	        this.complementarity = this.convertValues(source["complementarity"], ComplementarityResult);
	        this.penalties = this.convertValues(source["penalties"], PenaltyResult);
	        this.final = this.convertValues(source["final"], FinalScore);
	        this.warnings = source["warnings"];
	        this.pedigreeNotes = source["pedigreeNotes"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ComplementarityConfig {
	    maxBonusPoints: number;
	
	    static createFrom(source: any = {}) {
	        return new ComplementarityConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxBonusPoints = source["maxBonusPoints"];
	    }
	}
	
	
	
	export class DatasetInfo {
	    id: string;
	    packageId: string;
	    name: string;
	    version: string;
	    program: string;
	    release: string;
	    sha256: string;
	    isDemo: boolean;
	    origin: string;
	    licenseStatus: string;
	
	    static createFrom(source: any = {}) {
	        return new DatasetInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.packageId = source["packageId"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.program = source["program"];
	        this.release = source["release"];
	        this.sha256 = source["sha256"];
	        this.isDemo = source["isDemo"];
	        this.origin = source["origin"];
	        this.licenseStatus = source["licenseStatus"];
	    }
	}
	export class TraitMeta {
	    percentileSemantics: string;
	    populationMean?: number;
	    populationSd?: number;
	
	    static createFrom(source: any = {}) {
	        return new TraitMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.percentileSemantics = source["percentileSemantics"];
	        this.populationMean = source["populationMean"];
	        this.populationSd = source["populationSd"];
	    }
	}
	export class Evaluation {
	    id: string;
	    basis: Basis;
	    value: number;
	    accuracy?: number;
	    percentile?: number;
	    datasetId: string;
	    datasetName: string;
	    datasetVersion: string;
	    isDemo: boolean;
	    createdAt: string;
	    traitName: string;
	    direction: string;
	    idealMin?: number;
	    idealMax?: number;
	    meta: TraitMeta;
	
	    static createFrom(source: any = {}) {
	        return new Evaluation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.basis = this.convertValues(source["basis"], Basis);
	        this.value = source["value"];
	        this.accuracy = source["accuracy"];
	        this.percentile = source["percentile"];
	        this.datasetId = source["datasetId"];
	        this.datasetName = source["datasetName"];
	        this.datasetVersion = source["datasetVersion"];
	        this.isDemo = source["isDemo"];
	        this.createdAt = source["createdAt"];
	        this.traitName = source["traitName"];
	        this.direction = source["direction"];
	        this.idealMin = source["idealMin"];
	        this.idealMax = source["idealMax"];
	        this.meta = this.convertValues(source["meta"], TraitMeta);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Filters {
	    requireStock: boolean;
	    minDoses: number;
	    maxUnitCost?: number;
	    minAccuracy?: number;
	    excludeDirectAncestors: boolean;
	    ancestorGenerations: number;
	
	    static createFrom(source: any = {}) {
	        return new Filters(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requireStock = source["requireStock"];
	        this.minDoses = source["minDoses"];
	        this.maxUnitCost = source["maxUnitCost"];
	        this.minAccuracy = source["minAccuracy"];
	        this.excludeDirectAncestors = source["excludeDirectAncestors"];
	        this.ancestorGenerations = source["ancestorGenerations"];
	    }
	}
	
	export class LowAccuracyPenalty {
	    threshold: number;
	    maxPoints: number;
	
	    static createFrom(source: any = {}) {
	        return new LowAccuracyPenalty(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.threshold = source["threshold"];
	        this.maxPoints = source["maxPoints"];
	    }
	}
	export class ReferenceBasis {
	    traitCode: string;
	    evaluationType: string;
	    basis?: Basis;
	    candidates: number;
	    eligible: number;
	    damMatches: boolean;
	    reason: string;
	    alternatives: BasisCount[];
	
	    static createFrom(source: any = {}) {
	        return new ReferenceBasis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.traitCode = source["traitCode"];
	        this.evaluationType = source["evaluationType"];
	        this.basis = this.convertValues(source["basis"], Basis);
	        this.candidates = source["candidates"];
	        this.eligible = source["eligible"];
	        this.damMatches = source["damMatches"];
	        this.reason = source["reason"];
	        this.alternatives = this.convertValues(source["alternatives"], BasisCount);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PenaltyConfig {
	    lowAccuracy: LowAccuracyPenalty;
	
	    static createFrom(source: any = {}) {
	        return new PenaltyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.lowAccuracy = this.convertValues(source["lowAccuracy"], LowAccuracyPenalty);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TraitWeight {
	    traitCode: string;
	    evaluationType: string;
	    weight: number;
	    direction: string;
	    idealMin?: number;
	    idealMax?: number;
	    targetTolerance?: number;
	    required: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TraitWeight(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.traitCode = source["traitCode"];
	        this.evaluationType = source["evaluationType"];
	        this.weight = source["weight"];
	        this.direction = source["direction"];
	        this.idealMin = source["idealMin"];
	        this.idealMax = source["idealMax"];
	        this.targetTolerance = source["targetTolerance"];
	        this.required = source["required"];
	    }
	}
	export class Profile {
	    id: string;
	    name: string;
	    isDemo: boolean;
	    demoNotice: string;
	    description: string;
	    traits: TraitWeight[];
	    filters: Filters;
	    normalization: string;
	    minCoverage: number;
	    complementarity: ComplementarityConfig;
	    penalties: PenaltyConfig;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.isDemo = source["isDemo"];
	        this.demoNotice = source["demoNotice"];
	        this.description = source["description"];
	        this.traits = this.convertValues(source["traits"], TraitWeight);
	        this.filters = this.convertValues(source["filters"], Filters);
	        this.normalization = source["normalization"];
	        this.minCoverage = source["minCoverage"];
	        this.complementarity = this.convertValues(source["complementarity"], ComplementarityConfig);
	        this.penalties = this.convertValues(source["penalties"], PenaltyConfig);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PairSnapshot {
	    snapshotVersion: number;
	    engineVersion: string;
	    createdAt: string;
	    female: Animal;
	    femaleEvals: Evaluation[];
	    bull: Animal;
	    bullEvals: Evaluation[];
	    profile: Profile;
	    bases: ReferenceBasis[];
	    datasets: DatasetInfo[];
	    result: CandidateResult;
	    candidates: number;
	    runWarnings: string[];
	    isDemo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PairSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.snapshotVersion = source["snapshotVersion"];
	        this.engineVersion = source["engineVersion"];
	        this.createdAt = source["createdAt"];
	        this.female = this.convertValues(source["female"], Animal);
	        this.femaleEvals = this.convertValues(source["femaleEvals"], Evaluation);
	        this.bull = this.convertValues(source["bull"], Animal);
	        this.bullEvals = this.convertValues(source["bullEvals"], Evaluation);
	        this.profile = this.convertValues(source["profile"], Profile);
	        this.bases = this.convertValues(source["bases"], ReferenceBasis);
	        this.datasets = this.convertValues(source["datasets"], DatasetInfo);
	        this.result = this.convertValues(source["result"], CandidateResult);
	        this.candidates = source["candidates"];
	        this.runWarnings = source["runWarnings"];
	        this.isDemo = source["isDemo"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	export class RunOutput {
	    engineVersion: string;
	    createdAt: string;
	    female: Animal;
	    profile: Profile;
	    bases: ReferenceBasis[];
	    results: CandidateResult[];
	    top: string[];
	    datasets: DatasetInfo[];
	    errors: string[];
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new RunOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.engineVersion = source["engineVersion"];
	        this.createdAt = source["createdAt"];
	        this.female = this.convertValues(source["female"], Animal);
	        this.profile = this.convertValues(source["profile"], Profile);
	        this.bases = this.convertValues(source["bases"], ReferenceBasis);
	        this.results = this.convertValues(source["results"], CandidateResult);
	        this.top = source["top"];
	        this.datasets = this.convertValues(source["datasets"], DatasetInfo);
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	

}

export namespace genetics {
	
	export class CoefficientResult {
	    status: string;
	    value?: number;
	    method: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new CoefficientResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.value = source["value"];
	        this.method = source["method"];
	        this.message = source["message"];
	    }
	}
	export class GenerationStat {
	    generation: number;
	    expected: number;
	    known: number;
	    percent: number;
	
	    static createFrom(source: any = {}) {
	        return new GenerationStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generation = source["generation"];
	        this.expected = source["expected"];
	        this.known = source["known"];
	        this.percent = source["percent"];
	    }
	}
	export class Issue {
	    severity: string;
	    code: string;
	    animalId: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Issue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.severity = source["severity"];
	        this.code = source["code"];
	        this.animalId = source["animalId"];
	        this.message = source["message"];
	    }
	}
	export class Repeated {
	    id: string;
	    name: string;
	    occurrences: string[];
	
	    static createFrom(source: any = {}) {
	        return new Repeated(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.occurrences = source["occurrences"];
	    }
	}
	export class TreeNode {
	    id: string;
	    name: string;
	    registration: string;
	    sex: string;
	    breedName: string;
	    birthDate: string;
	    generation: number;
	    role: string;
	    known: boolean;
	    childId: string;
	    childName: string;
	    hasMore: boolean;
	    sire?: TreeNode;
	    dam?: TreeNode;
	
	    static createFrom(source: any = {}) {
	        return new TreeNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.registration = source["registration"];
	        this.sex = source["sex"];
	        this.breedName = source["breedName"];
	        this.birthDate = source["birthDate"];
	        this.generation = source["generation"];
	        this.role = source["role"];
	        this.known = source["known"];
	        this.childId = source["childId"];
	        this.childName = source["childName"];
	        this.hasMore = source["hasMore"];
	        this.sire = this.convertValues(source["sire"], TreeNode);
	        this.dam = this.convertValues(source["dam"], TreeNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace service {
	
	export class AnimalRef {
	    id: string;
	    name: string;
	    registration: string;
	    sex: string;
	    breedName: string;
	
	    static createFrom(source: any = {}) {
	        return new AnimalRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.registration = source["registration"];
	        this.sex = source["sex"];
	        this.breedName = source["breedName"];
	    }
	}
	export class AnimalDetail {
	    animal: domain.Animal;
	    breedName: string;
	    farmName: string;
	    father?: AnimalRef;
	    mother?: AnimalRef;
	    offspring: AnimalRef[];
	    semenDoses: number;
	
	    static createFrom(source: any = {}) {
	        return new AnimalDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.animal = this.convertValues(source["animal"], domain.Animal);
	        this.breedName = source["breedName"];
	        this.farmName = source["farmName"];
	        this.father = this.convertValues(source["father"], AnimalRef);
	        this.mother = this.convertValues(source["mother"], AnimalRef);
	        this.offspring = this.convertValues(source["offspring"], AnimalRef);
	        this.semenDoses = source["semenDoses"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AnimalInput {
	    farmId: string;
	    registration: string;
	    name: string;
	    sex: string;
	    breedId: string;
	    birthDate: string;
	    category: string;
	    status: string;
	    fatherId: string;
	    motherId: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new AnimalInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.farmId = source["farmId"];
	        this.registration = source["registration"];
	        this.name = source["name"];
	        this.sex = source["sex"];
	        this.breedId = source["breedId"];
	        this.birthDate = source["birthDate"];
	        this.category = source["category"];
	        this.status = source["status"];
	        this.fatherId = source["fatherId"];
	        this.motherId = source["motherId"];
	        this.notes = source["notes"];
	    }
	}
	export class AnimalPage {
	    items: domain.Animal[];
	    total: number;
	    page: number;
	
	    static createFrom(source: any = {}) {
	        return new AnimalPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], domain.Animal);
	        this.total = source["total"];
	        this.page = source["page"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AnimalQuery {
	    farmId: string;
	    breedId: string;
	    sex: string;
	    category: string;
	    status: string;
	    includeArchived: boolean;
	    search: string;
	    minAgeMonths?: number;
	    maxAgeMonths?: number;
	    sortBy: string;
	    sortDesc: boolean;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new AnimalQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.farmId = source["farmId"];
	        this.breedId = source["breedId"];
	        this.sex = source["sex"];
	        this.category = source["category"];
	        this.status = source["status"];
	        this.includeArchived = source["includeArchived"];
	        this.search = source["search"];
	        this.minAgeMonths = source["minAgeMonths"];
	        this.maxAgeMonths = source["maxAgeMonths"];
	        this.sortBy = source["sortBy"];
	        this.sortDesc = source["sortDesc"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	    }
	}
	
	export class AppInfo {
	    name: string;
	    version: string;
	    dataDir: string;
	    platform: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.dataDir = source["dataDir"];
	        this.platform = source["platform"];
	    }
	}
	export class BankRow {
	    id: string;
	    name: string;
	    registration: string;
	    breedName: string;
	    animalStatus: string;
	    doses: number;
	    lots: number;
	    suppliers: string[];
	    unitCost?: number;
	    stockStatus: string;
	    isDemo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BankRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.registration = source["registration"];
	        this.breedName = source["breedName"];
	        this.animalStatus = source["animalStatus"];
	        this.doses = source["doses"];
	        this.lots = source["lots"];
	        this.suppliers = source["suppliers"];
	        this.unitCost = source["unitCost"];
	        this.stockStatus = source["stockStatus"];
	        this.isDemo = source["isDemo"];
	    }
	}
	export class BankPage {
	    items: BankRow[];
	    total: number;
	    page: number;
	    totalDoses: number;
	    suppliers: string[];
	
	    static createFrom(source: any = {}) {
	        return new BankPage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], BankRow);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.totalDoses = source["totalDoses"];
	        this.suppliers = source["suppliers"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BankQuery {
	    search: string;
	    breedId: string;
	    supplier: string;
	    stockFilter: string;
	    includeArchived: boolean;
	    sortBy: string;
	    sortDesc: boolean;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new BankQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.search = source["search"];
	        this.breedId = source["breedId"];
	        this.supplier = source["supplier"];
	        this.stockFilter = source["stockFilter"];
	        this.includeArchived = source["includeArchived"];
	        this.sortBy = source["sortBy"];
	        this.sortDesc = source["sortDesc"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	    }
	}
	
	export class BullSheet {
	    detail?: AnimalDetail;
	    lots: domain.SemenInventory[];
	    totalDoses: number;
	    unitCost?: number;
	    movements: domain.SemenMovement[];
	    stockStatus: string;
	
	    static createFrom(source: any = {}) {
	        return new BullSheet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.detail = this.convertValues(source["detail"], AnimalDetail);
	        this.lots = this.convertValues(source["lots"], domain.SemenInventory);
	        this.totalDoses = source["totalDoses"];
	        this.unitCost = source["unitCost"];
	        this.movements = this.convertValues(source["movements"], domain.SemenMovement);
	        this.stockStatus = source["stockStatus"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CalculationSummary {
	    id: string;
	    runId: string;
	    bullId: string;
	    bullName: string;
	    profileName: string;
	    status: string;
	    score?: number;
	    rank: number;
	    engineVersion: string;
	    createdAt: string;
	    isDemo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CalculationSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.runId = source["runId"];
	        this.bullId = source["bullId"];
	        this.bullName = source["bullName"];
	        this.profileName = source["profileName"];
	        this.status = source["status"];
	        this.score = source["score"];
	        this.rank = source["rank"];
	        this.engineVersion = source["engineVersion"];
	        this.createdAt = source["createdAt"];
	        this.isDemo = source["isDemo"];
	    }
	}
	export class CalculationView {
	    id: string;
	    runId: string;
	    femaleId: string;
	    bullId: string;
	    status: string;
	    score?: number;
	    engineVersion: string;
	    createdAt: string;
	    isDemo: boolean;
	    snapshot: number[];
	
	    static createFrom(source: any = {}) {
	        return new CalculationView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.runId = source["runId"];
	        this.femaleId = source["femaleId"];
	        this.bullId = source["bullId"];
	        this.status = source["status"];
	        this.score = source["score"];
	        this.engineVersion = source["engineVersion"];
	        this.createdAt = source["createdAt"];
	        this.isDemo = source["isDemo"];
	        this.snapshot = source["snapshot"];
	    }
	}
	export class CommonAncestorView {
	    id: string;
	    generationsFromA: number[];
	    generationsFromB: number[];
	    depth: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new CommonAncestorView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.generationsFromA = source["generationsFromA"];
	        this.generationsFromB = source["generationsFromB"];
	        this.depth = source["depth"];
	        this.name = source["name"];
	    }
	}
	export class CommonAncestorsResult {
	    ancestors: CommonAncestorView[];
	    relationship: genetics.CoefficientResult;
	
	    static createFrom(source: any = {}) {
	        return new CommonAncestorsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ancestors = this.convertValues(source["ancestors"], CommonAncestorView);
	        this.relationship = this.convertValues(source["relationship"], genetics.CoefficientResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EvaluationInput {
	    animalId: string;
	    datasetId: string;
	    evaluationProgram: string;
	    evaluationRelease: string;
	    breedCode: string;
	    traitCode: string;
	    evaluationType: string;
	    value: number;
	    unit: string;
	    accuracy?: number;
	    percentile?: number;
	    geneticBase: string;
	    sourceReference: string;
	    evaluatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new EvaluationInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.animalId = source["animalId"];
	        this.datasetId = source["datasetId"];
	        this.evaluationProgram = source["evaluationProgram"];
	        this.evaluationRelease = source["evaluationRelease"];
	        this.breedCode = source["breedCode"];
	        this.traitCode = source["traitCode"];
	        this.evaluationType = source["evaluationType"];
	        this.value = source["value"];
	        this.unit = source["unit"];
	        this.accuracy = source["accuracy"];
	        this.percentile = source["percentile"];
	        this.geneticBase = source["geneticBase"];
	        this.sourceReference = source["sourceReference"];
	        this.evaluatedAt = source["evaluatedAt"];
	    }
	}
	export class FarmInput {
	    name: string;
	    ownerName: string;
	    city: string;
	    state: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new FarmInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.ownerName = source["ownerName"];
	        this.city = source["city"];
	        this.state = source["state"];
	        this.notes = source["notes"];
	    }
	}
	export class ImportValidation {
	    report: datasets.Report;
	    licenseStatus: string;
	    duplicate: boolean;
	    canInstall: boolean;
	    notice: string;
	    manifest: datasets.Manifest;
	
	    static createFrom(source: any = {}) {
	        return new ImportValidation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.report = this.convertValues(source["report"], datasets.Report);
	        this.licenseStatus = source["licenseStatus"];
	        this.duplicate = source["duplicate"];
	        this.canInstall = source["canInstall"];
	        this.notice = source["notice"];
	        this.manifest = this.convertValues(source["manifest"], datasets.Manifest);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ImportInspection {
	    kind: string;
	    path: string;
	    columns: string[];
	    sampleRows: string[][];
	    rowCount: number;
	    suggestedDecimalComma: boolean;
	    suggestedIdColumn: string;
	    suggestedNameColumn: string;
	    manifest?: datasets.Manifest;
	    traits: datasets.TraitDef[];
	    readme: string;
	    licenseText: string;
	    validation?: ImportValidation;
	
	    static createFrom(source: any = {}) {
	        return new ImportInspection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.path = source["path"];
	        this.columns = source["columns"];
	        this.sampleRows = source["sampleRows"];
	        this.rowCount = source["rowCount"];
	        this.suggestedDecimalComma = source["suggestedDecimalComma"];
	        this.suggestedIdColumn = source["suggestedIdColumn"];
	        this.suggestedNameColumn = source["suggestedNameColumn"];
	        this.manifest = this.convertValues(source["manifest"], datasets.Manifest);
	        this.traits = this.convertValues(source["traits"], datasets.TraitDef);
	        this.readme = source["readme"];
	        this.licenseText = source["licenseText"];
	        this.validation = this.convertValues(source["validation"], ImportValidation);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ImportRequest {
	    path: string;
	    manifest: datasets.Manifest;
	    traits: datasets.TraitDef[];
	    licenseText: string;
	    unmatchedPolicy: string;
	    farmId: string;
	    acceptResponsibility: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ImportRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.manifest = this.convertValues(source["manifest"], datasets.Manifest);
	        this.traits = this.convertValues(source["traits"], datasets.TraitDef);
	        this.licenseText = source["licenseText"];
	        this.unmatchedPolicy = source["unmatchedPolicy"];
	        this.farmId = source["farmId"];
	        this.acceptResponsibility = source["acceptResponsibility"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class LibrarySummary {
	    id: string;
	    packageId: string;
	    name: string;
	    version: string;
	    origin: string;
	    species: string;
	    breeds: string[];
	    purposes: string[];
	    organization: string;
	    sourceUrl: string;
	    program: string;
	    release: string;
	    licenseName: string;
	    licenseUrl: string;
	    licenseStatus: string;
	    redistributionAllowed: boolean;
	    active: boolean;
	    isDemo: boolean;
	    recordCount: number;
	    traitCount: number;
	    fileSha256: string;
	    installedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new LibrarySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.packageId = source["packageId"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.origin = source["origin"];
	        this.species = source["species"];
	        this.breeds = source["breeds"];
	        this.purposes = source["purposes"];
	        this.organization = source["organization"];
	        this.sourceUrl = source["sourceUrl"];
	        this.program = source["program"];
	        this.release = source["release"];
	        this.licenseName = source["licenseName"];
	        this.licenseUrl = source["licenseUrl"];
	        this.licenseStatus = source["licenseStatus"];
	        this.redistributionAllowed = source["redistributionAllowed"];
	        this.active = source["active"];
	        this.isDemo = source["isDemo"];
	        this.recordCount = source["recordCount"];
	        this.traitCount = source["traitCount"];
	        this.fileSha256 = source["fileSha256"];
	        this.installedAt = source["installedAt"];
	    }
	}
	export class InstallResult {
	    library: LibrarySummary;
	    report: datasets.Report;
	    evaluationsInserted: number;
	    animalsCreated: number;
	    recordsSkipped: number;
	
	    static createFrom(source: any = {}) {
	        return new InstallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.library = this.convertValues(source["library"], LibrarySummary);
	        this.report = this.convertValues(source["report"], datasets.Report);
	        this.evaluationsInserted = source["evaluationsInserted"];
	        this.animalsCreated = source["animalsCreated"];
	        this.recordsSkipped = source["recordsSkipped"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LibraryDetail {
	    id: string;
	    packageId: string;
	    name: string;
	    version: string;
	    origin: string;
	    species: string;
	    breeds: string[];
	    purposes: string[];
	    organization: string;
	    sourceUrl: string;
	    program: string;
	    release: string;
	    licenseName: string;
	    licenseUrl: string;
	    licenseStatus: string;
	    redistributionAllowed: boolean;
	    active: boolean;
	    isDemo: boolean;
	    recordCount: number;
	    traitCount: number;
	    fileSha256: string;
	    installedAt: string;
	    manifest: number[];
	    readme: string;
	    licenseText: string;
	    traits: domain.Trait[];
	    usedByCalculations: number;
	    responsibilityNotice: string;
	
	    static createFrom(source: any = {}) {
	        return new LibraryDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.packageId = source["packageId"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.origin = source["origin"];
	        this.species = source["species"];
	        this.breeds = source["breeds"];
	        this.purposes = source["purposes"];
	        this.organization = source["organization"];
	        this.sourceUrl = source["sourceUrl"];
	        this.program = source["program"];
	        this.release = source["release"];
	        this.licenseName = source["licenseName"];
	        this.licenseUrl = source["licenseUrl"];
	        this.licenseStatus = source["licenseStatus"];
	        this.redistributionAllowed = source["redistributionAllowed"];
	        this.active = source["active"];
	        this.isDemo = source["isDemo"];
	        this.recordCount = source["recordCount"];
	        this.traitCount = source["traitCount"];
	        this.fileSha256 = source["fileSha256"];
	        this.installedAt = source["installedAt"];
	        this.manifest = source["manifest"];
	        this.readme = source["readme"];
	        this.licenseText = source["licenseText"];
	        this.traits = this.convertValues(source["traits"], domain.Trait);
	        this.usedByCalculations = source["usedByCalculations"];
	        this.responsibilityNotice = source["responsibilityNotice"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class LotInfoInput {
	    supplier: string;
	    batch: string;
	    unitCost?: number;
	    semenType: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new LotInfoInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supplier = source["supplier"];
	        this.batch = source["batch"];
	        this.unitCost = source["unitCost"];
	        this.semenType = source["semenType"];
	        this.notes = source["notes"];
	    }
	}
	export class MeasurementInput {
	    animalId: string;
	    traitCode: string;
	    value: number;
	    unit: string;
	    measuredAt: string;
	    source: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new MeasurementInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.animalId = source["animalId"];
	        this.traitCode = source["traitCode"];
	        this.value = source["value"];
	        this.unit = source["unit"];
	        this.measuredAt = source["measuredAt"];
	        this.source = source["source"];
	        this.notes = source["notes"];
	    }
	}
	export class MinimalAncestorInput {
	    childId: string;
	    role: string;
	    name: string;
	    registration: string;
	    breedId: string;
	
	    static createFrom(source: any = {}) {
	        return new MinimalAncestorInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.childId = source["childId"];
	        this.role = source["role"];
	        this.name = source["name"];
	        this.registration = source["registration"];
	        this.breedId = source["breedId"];
	    }
	}
	export class MovementInput {
	    inventoryId: string;
	    quantity: number;
	    reason: string;
	    note: string;
	    matingId: string;
	
	    static createFrom(source: any = {}) {
	        return new MovementInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.inventoryId = source["inventoryId"];
	        this.quantity = source["quantity"];
	        this.reason = source["reason"];
	        this.note = source["note"];
	        this.matingId = source["matingId"];
	    }
	}
	export class PedigreeView {
	    animalId: string;
	    generations: number;
	    tree?: genetics.TreeNode;
	    depth: number;
	    stats: genetics.GenerationStat[];
	    repeated: genetics.Repeated[];
	    issues: genetics.Issue[];
	    inbreeding: genetics.CoefficientResult;
	
	    static createFrom(source: any = {}) {
	        return new PedigreeView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.animalId = source["animalId"];
	        this.generations = source["generations"];
	        this.tree = this.convertValues(source["tree"], genetics.TreeNode);
	        this.depth = source["depth"];
	        this.stats = this.convertValues(source["stats"], genetics.GenerationStat);
	        this.repeated = this.convertValues(source["repeated"], genetics.Repeated);
	        this.issues = this.convertValues(source["issues"], genetics.Issue);
	        this.inbreeding = this.convertValues(source["inbreeding"], genetics.CoefficientResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Provenance {
	    kindCode: string;
	    kindLabel: string;
	    genetic: boolean;
	    traitCode: string;
	    traitName: string;
	    datasetId: string;
	    datasetName: string;
	    datasetVersion: string;
	    program: string;
	    release: string;
	    geneticBase: string;
	    sourceReference: string;
	    accuracy?: number;
	    source: string;
	    date: string;
	    isDemo: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Provenance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kindCode = source["kindCode"];
	        this.kindLabel = source["kindLabel"];
	        this.genetic = source["genetic"];
	        this.traitCode = source["traitCode"];
	        this.traitName = source["traitName"];
	        this.datasetId = source["datasetId"];
	        this.datasetName = source["datasetName"];
	        this.datasetVersion = source["datasetVersion"];
	        this.program = source["program"];
	        this.release = source["release"];
	        this.geneticBase = source["geneticBase"];
	        this.sourceReference = source["sourceReference"];
	        this.accuracy = source["accuracy"];
	        this.source = source["source"];
	        this.date = source["date"];
	        this.isDemo = source["isDemo"];
	    }
	}
	export class RunRequest {
	    femaleId: string;
	    bullIds: string[];
	    profile: engine.Profile;
	    save: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RunRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.femaleId = source["femaleId"];
	        this.bullIds = source["bullIds"];
	        this.profile = this.convertValues(source["profile"], engine.Profile);
	        this.save = source["save"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RunResponse {
	    runId: string;
	    saved: boolean;
	    matingIds: Record<string, string>;
	    output: engine.RunOutput;
	    snapshots: Record<string, engine.PairSnapshot>;
	
	    static createFrom(source: any = {}) {
	        return new RunResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.runId = source["runId"];
	        this.saved = source["saved"];
	        this.matingIds = source["matingIds"];
	        this.output = this.convertValues(source["output"], engine.RunOutput);
	        this.snapshots = this.convertValues(source["snapshots"], engine.PairSnapshot, true);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SemenInput {
	    bullId: string;
	    supplier: string;
	    batch: string;
	    quantity: number;
	    unitCost?: number;
	    semenType: string;
	    notes: string;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new SemenInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bullId = source["bullId"];
	        this.supplier = source["supplier"];
	        this.batch = source["batch"];
	        this.quantity = source["quantity"];
	        this.unitCost = source["unitCost"];
	        this.semenType = source["semenType"];
	        this.notes = source["notes"];
	        this.reason = source["reason"];
	    }
	}
	export class TraitItem {
	    id: string;
	    source: string;
	    traitCode: string;
	    traitName: string;
	    defined: boolean;
	    category: string;
	    valueKind: string;
	    unit: string;
	    value: number;
	    accuracy?: number;
	    percentile?: number;
	    notes: string;
	    optimizationDirection: string;
	    provenance: Provenance;
	
	    static createFrom(source: any = {}) {
	        return new TraitItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.source = source["source"];
	        this.traitCode = source["traitCode"];
	        this.traitName = source["traitName"];
	        this.defined = source["defined"];
	        this.category = source["category"];
	        this.valueKind = source["valueKind"];
	        this.unit = source["unit"];
	        this.value = source["value"];
	        this.accuracy = source["accuracy"];
	        this.percentile = source["percentile"];
	        this.notes = source["notes"];
	        this.optimizationDirection = source["optimizationDirection"];
	        this.provenance = this.convertValues(source["provenance"], Provenance);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TraitGroup {
	    category: string;
	    items: TraitItem[];
	
	    static createFrom(source: any = {}) {
	        return new TraitGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.items = this.convertValues(source["items"], TraitItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TraitInput {
	    code: string;
	    name: string;
	    description: string;
	    category: string;
	    purpose: string;
	    unit: string;
	    dataType: string;
	    evaluationType: string;
	    optimizationDirection: string;
	    idealMin?: number;
	    idealMax?: number;
	    higherIsBetter?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TraitInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.purpose = source["purpose"];
	        this.unit = source["unit"];
	        this.dataType = source["dataType"];
	        this.evaluationType = source["evaluationType"];
	        this.optimizationDirection = source["optimizationDirection"];
	        this.idealMin = source["idealMin"];
	        this.idealMax = source["idealMax"];
	        this.higherIsBetter = source["higherIsBetter"];
	    }
	}
	
	export class TraitQuery {
	    kind: string;
	    category: string;
	    search: string;
	
	    static createFrom(source: any = {}) {
	        return new TraitQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.category = source["category"];
	        this.search = source["search"];
	    }
	}

}

