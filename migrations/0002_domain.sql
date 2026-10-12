-- Domain schema. farm.db holds the user's private data only; scientific
-- datasets are installed separately and referenced through `datasets`.
-- IDs are UUID v4 strings. Timestamps are RFC 3339 UTC; dates are YYYY-MM-DD.
-- is_demo = 1 marks DEMO / SYNTHETIC rows so they are never mistaken for real data.

CREATE TABLE farms (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL CHECK (length(trim(name)) > 0),
    owner_name TEXT,
    city       TEXT,
    state      TEXT,
    notes      TEXT NOT NULL DEFAULT '',
    is_demo    INTEGER NOT NULL DEFAULT 0 CHECK (is_demo IN (0, 1)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE breeds (
    id      TEXT PRIMARY KEY,
    code    TEXT NOT NULL UNIQUE CHECK (length(trim(code)) > 0),
    name    TEXT NOT NULL,
    species TEXT NOT NULL,
    purpose TEXT NOT NULL CHECK (purpose IN ('beef', 'dairy', 'dual')),
    active  INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1))
);

-- Minimal local catalogue so the UI can work offline. This is a list of names,
-- not genetic knowledge; evaluations always come from datasets.
INSERT INTO breeds (id, code, name, species, purpose) VALUES
    ('breed-nelore',   'NEL', 'Nelore',            'Bos indicus',  'beef'),
    ('breed-gir',      'GIR', 'Gir',               'Bos indicus',  'dairy'),
    ('breed-angus',    'ANG', 'Angus',             'Bos taurus',   'beef'),
    ('breed-hereford', 'HER', 'Hereford',          'Bos taurus',   'beef'),
    ('breed-holstein', 'HOL', 'Holandesa',         'Bos taurus',   'dairy'),
    ('breed-jersey',   'JER', 'Jersey',            'Bos taurus',   'dairy'),
    ('breed-crossbred','CRZ', 'Cruzado / mestiço', 'Bos spp.',     'dual');

-- Installed scientific libraries. Extended by the library module.
CREATE TABLE datasets (
    id           TEXT PRIMARY KEY,
    package_id   TEXT NOT NULL,
    name         TEXT NOT NULL,
    version      TEXT NOT NULL,
    is_demo      INTEGER NOT NULL DEFAULT 0 CHECK (is_demo IN (0, 1)),
    installed_at TEXT NOT NULL,
    UNIQUE (package_id, version)
);

CREATE TABLE animals (
    id           TEXT PRIMARY KEY,
    farm_id      TEXT NOT NULL REFERENCES farms(id) ON DELETE RESTRICT,
    registration TEXT,
    name         TEXT NOT NULL DEFAULT '',
    sex          TEXT NOT NULL CHECK (sex IN ('male', 'female')),
    breed_id     TEXT REFERENCES breeds(id) ON DELETE RESTRICT,
    birth_date   TEXT CHECK (birth_date IS NULL OR birth_date GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]'),
    category     TEXT NOT NULL CHECK (category IN ('calf', 'heifer', 'cow', 'steer', 'bull', 'other')),
    status       TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived', 'sold', 'deceased')),
    father_id    TEXT REFERENCES animals(id) ON DELETE RESTRICT,
    mother_id    TEXT REFERENCES animals(id) ON DELETE RESTRICT,
    notes        TEXT NOT NULL DEFAULT '',
    is_demo      INTEGER NOT NULL DEFAULT 0 CHECK (is_demo IN (0, 1)),
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL,
    CHECK (father_id IS NULL OR father_id <> id),
    CHECK (mother_id IS NULL OR mother_id <> id),
    CHECK (father_id IS NULL OR mother_id IS NULL OR father_id <> mother_id)
);
CREATE INDEX idx_animals_farm      ON animals(farm_id, status);
CREATE INDEX idx_animals_father    ON animals(father_id);
CREATE INDEX idx_animals_mother    ON animals(mother_id);
CREATE INDEX idx_animals_breed     ON animals(breed_id);
CREATE INDEX idx_animals_name      ON animals(name COLLATE NOCASE);
CREATE UNIQUE INDEX uq_animals_registration ON animals(farm_id, registration)
    WHERE registration IS NOT NULL AND registration <> '';

CREATE TABLE animal_measurements (
    id          TEXT PRIMARY KEY,
    animal_id   TEXT NOT NULL REFERENCES animals(id) ON DELETE CASCADE,
    trait_code  TEXT NOT NULL CHECK (length(trim(trait_code)) > 0),
    value       REAL NOT NULL,
    unit        TEXT NOT NULL DEFAULT '',
    measured_at TEXT NOT NULL,
    source      TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    is_demo     INTEGER NOT NULL DEFAULT 0 CHECK (is_demo IN (0, 1)),
    created_at  TEXT NOT NULL
);
CREATE INDEX idx_measurements_animal ON animal_measurements(animal_id, trait_code, measured_at);

CREATE TABLE genetic_evaluations (
    id                 TEXT PRIMARY KEY,
    animal_id          TEXT NOT NULL REFERENCES animals(id) ON DELETE CASCADE,
    dataset_id         TEXT REFERENCES datasets(id) ON DELETE RESTRICT,
    evaluation_program TEXT NOT NULL,
    evaluation_release TEXT NOT NULL,
    breed_code         TEXT NOT NULL DEFAULT '',
    trait_code         TEXT NOT NULL CHECK (length(trim(trait_code)) > 0),
    evaluation_type    TEXT NOT NULL CHECK (evaluation_type IN ('DEP', 'EPD', 'EBV', 'PTA', 'INDEX')),
    value              REAL NOT NULL,
    unit               TEXT NOT NULL DEFAULT '',
    accuracy           REAL CHECK (accuracy IS NULL OR (accuracy >= 0 AND accuracy <= 1)),
    percentile         REAL CHECK (percentile IS NULL OR (percentile >= 0 AND percentile <= 100)),
    genetic_base       TEXT,
    source_reference   TEXT,
    evaluated_at       TEXT,
    is_demo            INTEGER NOT NULL DEFAULT 0 CHECK (is_demo IN (0, 1)),
    created_at         TEXT NOT NULL
);
CREATE INDEX idx_evaluations_animal  ON genetic_evaluations(animal_id, trait_code);
CREATE INDEX idx_evaluations_dataset ON genetic_evaluations(dataset_id);

CREATE TABLE semen_inventory (
    id         TEXT PRIMARY KEY,
    bull_id    TEXT NOT NULL REFERENCES animals(id) ON DELETE RESTRICT,
    supplier   TEXT NOT NULL DEFAULT '',
    batch      TEXT NOT NULL DEFAULT '',
    quantity   INTEGER NOT NULL CHECK (quantity >= 0),
    unit_cost  REAL CHECK (unit_cost IS NULL OR unit_cost >= 0),
    semen_type TEXT NOT NULL DEFAULT 'conventional',
    notes      TEXT NOT NULL DEFAULT '',
    is_demo    INTEGER NOT NULL DEFAULT 0 CHECK (is_demo IN (0, 1)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX idx_semen_bull ON semen_inventory(bull_id);

CREATE TABLE matings (
    id                   TEXT PRIMARY KEY,
    farm_id              TEXT NOT NULL REFERENCES farms(id) ON DELETE RESTRICT,
    female_id            TEXT NOT NULL REFERENCES animals(id) ON DELETE RESTRICT,
    bull_id              TEXT NOT NULL REFERENCES animals(id) ON DELETE RESTRICT,
    status               TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'planned', 'confirmed', 'done', 'cancelled')),
    score                REAL,
    engine_version       TEXT NOT NULL,
    calculation_snapshot TEXT NOT NULL CHECK (json_valid(calculation_snapshot)),
    created_at           TEXT NOT NULL,
    CHECK (female_id <> bull_id)
);
CREATE INDEX idx_matings_female ON matings(female_id);
CREATE INDEX idx_matings_bull   ON matings(bull_id);
