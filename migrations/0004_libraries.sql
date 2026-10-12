-- Dataset packages ("bibliotecas"). Each installed version is a separate row so
-- historical calculations keep pointing at the exact data they used.
-- The original package is kept on disk (libraries/<id>.genlib) and its SHA-256 recorded.
ALTER TABLE datasets ADD COLUMN origin TEXT NOT NULL DEFAULT 'genlib' CHECK (origin IN ('genlib', 'csv', 'json'));
ALTER TABLE datasets ADD COLUMN manifest TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(manifest));
ALTER TABLE datasets ADD COLUMN species TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN breeds TEXT NOT NULL DEFAULT '[]' CHECK (json_valid(breeds));
ALTER TABLE datasets ADD COLUMN purposes TEXT NOT NULL DEFAULT '[]' CHECK (json_valid(purposes));
ALTER TABLE datasets ADD COLUMN organization TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN source_url TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN program TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN release TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN license_name TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN license_url TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN license_status TEXT NOT NULL DEFAULT 'unknown' CHECK (license_status IN ('unknown', 'no_redistribute', 'redistributable'));
ALTER TABLE datasets ADD COLUMN redistribution_allowed INTEGER NOT NULL DEFAULT 0 CHECK (redistribution_allowed IN (0, 1));
ALTER TABLE datasets ADD COLUMN readme TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN license_text TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN record_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE datasets ADD COLUMN trait_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE datasets ADD COLUMN active INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1));
ALTER TABLE datasets ADD COLUMN file_name TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN file_sha256 TEXT NOT NULL DEFAULT '';
ALTER TABLE datasets ADD COLUMN responsibility_accepted INTEGER NOT NULL DEFAULT 0 CHECK (responsibility_accepted IN (0, 1));

-- Which dataset versions each saved calculation used (filled by the mating engine).
CREATE TABLE calculation_datasets (
    mating_id  TEXT NOT NULL REFERENCES matings(id) ON DELETE CASCADE,
    dataset_id TEXT NOT NULL REFERENCES datasets(id) ON DELETE RESTRICT,
    PRIMARY KEY (mating_id, dataset_id)
);
CREATE INDEX idx_calc_datasets_dataset ON calculation_datasets(dataset_id);
