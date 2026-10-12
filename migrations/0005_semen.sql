-- Semen types are configurable; two are built in.
CREATE TABLE semen_types (
    code       TEXT PRIMARY KEY CHECK (code GLOB '[a-z]*' AND length(code) <= 40),
    name       TEXT NOT NULL CHECK (length(trim(name)) > 0),
    is_builtin INTEGER NOT NULL DEFAULT 0 CHECK (is_builtin IN (0, 1))
);
INSERT INTO semen_types (code, name, is_builtin) VALUES ('conventional', 'Convencional', 1), ('sexed', 'Sexado', 1);

-- Every change of stock is recorded. balance_after can never be negative.
CREATE TABLE semen_movements (
    id             TEXT PRIMARY KEY,
    inventory_id   TEXT NOT NULL REFERENCES semen_inventory(id) ON DELETE RESTRICT,
    bull_id        TEXT NOT NULL REFERENCES animals(id) ON DELETE RESTRICT,
    movement_type  TEXT NOT NULL CHECK (movement_type IN ('entry', 'withdrawal', 'adjustment')),
    delta          INTEGER NOT NULL CHECK (delta <> 0),
    balance_after  INTEGER NOT NULL CHECK (balance_after >= 0),
    reason         TEXT NOT NULL DEFAULT '',
    note           TEXT NOT NULL DEFAULT '',
    mating_id      TEXT REFERENCES matings(id) ON DELETE SET NULL,
    is_demo        INTEGER NOT NULL DEFAULT 0 CHECK (is_demo IN (0, 1)),
    created_at     TEXT NOT NULL,
    CHECK ((movement_type = 'entry' AND delta > 0) OR (movement_type = 'withdrawal' AND delta < 0) OR movement_type = 'adjustment')
);
CREATE INDEX idx_semen_mov_bull ON semen_movements(bull_id, created_at);
CREATE INDEX idx_semen_mov_inventory ON semen_movements(inventory_id, created_at);
