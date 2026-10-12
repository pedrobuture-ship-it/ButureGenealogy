-- A calculation run groups the pair snapshots produced by one "calculate" click.
CREATE TABLE calculation_runs (
    id             TEXT PRIMARY KEY,
    farm_id        TEXT NOT NULL REFERENCES farms(id) ON DELETE RESTRICT,
    female_id      TEXT NOT NULL REFERENCES animals(id) ON DELETE RESTRICT,
    profile_name   TEXT NOT NULL,
    engine_version TEXT NOT NULL,
    candidate_count INTEGER NOT NULL,
    summary        TEXT NOT NULL CHECK (json_valid(summary)), -- bases, warnings, exclusions, top
    is_demo        INTEGER NOT NULL DEFAULT 0 CHECK (is_demo IN (0, 1)),
    created_at     TEXT NOT NULL
);
CREATE INDEX idx_runs_female ON calculation_runs(female_id, created_at);

ALTER TABLE matings ADD COLUMN run_id TEXT REFERENCES calculation_runs(id) ON DELETE SET NULL;
ALTER TABLE matings ADD COLUMN is_demo INTEGER NOT NULL DEFAULT 0 CHECK (is_demo IN (0, 1));
CREATE INDEX idx_matings_run ON matings(run_id);

-- User-saved breeding objectives. Built-in and DEMO profiles live in code.
CREATE TABLE mating_profiles (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE CHECK (length(trim(name)) > 0),
    config     TEXT NOT NULL CHECK (json_valid(config)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
