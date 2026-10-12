-- Metadata about the database itself. Domain tables arrive in later migrations.
CREATE TABLE app_meta (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO app_meta (key, value) VALUES ('created_by', 'genetica');
