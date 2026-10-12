-- Generic trait catalogue. No breed-specific rules live in code: a trait is
-- data. `evaluation_type` says what kind of value the trait describes, and
-- PHENOTYPE / DEP / EPD / EBV / PTA / INDEX are never interchangeable.
-- Built-in rows are neutral names only: their direction is INFORMATION_ONLY,
-- because "better" depends on the breeding objective, which belongs to a
-- profile or to a dataset, not to the application.
CREATE TABLE traits (
    id                     TEXT PRIMARY KEY,
    code                   TEXT NOT NULL CHECK (length(trim(code)) > 0),
    name                   TEXT NOT NULL CHECK (length(trim(name)) > 0),
    description            TEXT NOT NULL DEFAULT '',
    category               TEXT NOT NULL CHECK (category IN
        ('growth','reproduction','maternal','carcass','dairy','conformation','health','temperament','efficiency','other')),
    purpose                TEXT NOT NULL DEFAULT 'any' CHECK (purpose IN ('beef','dairy','dual','any')),
    unit                   TEXT NOT NULL DEFAULT '',
    data_type              TEXT NOT NULL DEFAULT 'number' CHECK (data_type IN ('number','integer','score')),
    evaluation_type        TEXT NOT NULL CHECK (evaluation_type IN ('PHENOTYPE','DEP','EPD','EBV','PTA','INDEX')),
    optimization_direction TEXT NOT NULL DEFAULT 'INFORMATION_ONLY'
        CHECK (optimization_direction IN ('MAXIMIZE','MINIMIZE','TARGET_RANGE','INFORMATION_ONLY')),
    ideal_min              REAL,
    ideal_max              REAL,
    higher_is_better       INTEGER CHECK (higher_is_better IS NULL OR higher_is_better IN (0, 1)),
    source_dataset         TEXT NOT NULL DEFAULT '',
    metadata               TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(metadata)),
    is_builtin             INTEGER NOT NULL DEFAULT 0 CHECK (is_builtin IN (0, 1)),
    created_at             TEXT NOT NULL,
    UNIQUE (code, evaluation_type, source_dataset),
    CHECK (optimization_direction <> 'TARGET_RANGE' OR (ideal_min IS NOT NULL AND ideal_max IS NOT NULL)),
    CHECK (ideal_min IS NULL OR ideal_max IS NULL OR ideal_min <= ideal_max)
);
CREATE INDEX idx_traits_category ON traits(category, name);

INSERT INTO traits (id, code, name, description, category, purpose, unit, data_type, evaluation_type, is_builtin, created_at) VALUES
 ('builtin:PHENOTYPE:BODY_WEIGHT',   'BODY_WEIGHT',   'Peso corporal',               'Peso vivo medido no animal.',                         'growth',       'any',   'kg', 'number',  'PHENOTYPE', 1, '2026-01-01T00:00:00Z'),
 ('builtin:PHENOTYPE:BIRTH_WEIGHT',  'BIRTH_WEIGHT',  'Peso ao nascer (medido)',     'Peso do bezerro ao nascimento, observado.',            'growth',       'any',   'kg', 'number',  'PHENOTYPE', 1, '2026-01-01T00:00:00Z'),
 ('builtin:PHENOTYPE:HIP_HEIGHT',    'HIP_HEIGHT',    'Altura de garupa',            'Altura medida na garupa.',                            'conformation', 'any',   'cm', 'number',  'PHENOTYPE', 1, '2026-01-01T00:00:00Z'),
 ('builtin:PHENOTYPE:SCROTAL_CIRC',  'SCROTAL_CIRC',  'Perímetro escrotal',          'Circunferência escrotal medida em machos.',           'reproduction', 'any',   'cm', 'number',  'PHENOTYPE', 1, '2026-01-01T00:00:00Z'),
 ('builtin:PHENOTYPE:BCS',           'BCS',           'Escore de condição corporal', 'Escore subjetivo da condição corporal (escala do avaliador).', 'efficiency', 'any', '',   'score',   'PHENOTYPE', 1, '2026-01-01T00:00:00Z'),
 ('builtin:PHENOTYPE:MILK_YIELD',    'MILK_YIELD',    'Produção de leite',           'Produção de leite medida (por dia ou por lactação, conforme a origem).', 'dairy', 'dairy', 'kg', 'number', 'PHENOTYPE', 1, '2026-01-01T00:00:00Z'),
 ('builtin:PHENOTYPE:TEMPERAMENT',   'TEMPERAMENT',   'Temperamento (escore)',       'Escore de temperamento (escala do avaliador).',       'temperament',  'any',   '',   'score',   'PHENOTYPE', 1, '2026-01-01T00:00:00Z'),
 ('builtin:DEP:BW',   'BW',   'Peso ao nascer',          'Característica genética: peso ao nascer.',        'growth',       'beef', 'kg', 'number', 'DEP', 1, '2026-01-01T00:00:00Z'),
 ('builtin:DEP:WW',   'WW',   'Peso à desmama',          'Característica genética: peso à desmama.',        'growth',       'beef', 'kg', 'number', 'DEP', 1, '2026-01-01T00:00:00Z'),
 ('builtin:DEP:YW',   'YW',   'Peso ao sobreano',        'Característica genética: peso ao sobreano.',      'growth',       'beef', 'kg', 'number', 'DEP', 1, '2026-01-01T00:00:00Z'),
 ('builtin:DEP:MILK', 'MILK', 'Habilidade materna (leite)', 'Característica genética: efeito materno sobre a desmama.', 'maternal', 'beef', 'kg', 'number', 'DEP', 1, '2026-01-01T00:00:00Z');
