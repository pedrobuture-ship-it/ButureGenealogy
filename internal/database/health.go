package database

import "context"

// Health is a snapshot of the database state.
type Health struct {
	OK            bool   `json:"ok"`
	Path          string `json:"path"`
	SchemaVersion int    `json:"schemaVersion"`
	JournalMode   string `json:"journalMode"`
	ForeignKeys   bool   `json:"foreignKeys"`
	QuickCheck    string `json:"quickCheck"`
	Error         string `json:"error,omitempty"`
}

// Health pings the database and runs PRAGMA quick_check.
func (d *DB) Health(ctx context.Context) Health {
	h := Health{Path: d.path}
	if err := d.PingContext(ctx); err != nil {
		h.Error = err.Error()
		return h
	}
	v, err := d.SchemaVersion(ctx)
	if err != nil {
		h.Error = err.Error()
		return h
	}
	h.SchemaVersion = v
	_ = d.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&h.JournalMode)
	var fk int
	_ = d.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&fk)
	h.ForeignKeys = fk == 1
	if err := d.QueryRowContext(ctx, `PRAGMA quick_check`).Scan(&h.QuickCheck); err != nil {
		h.Error = err.Error()
		return h
	}
	h.OK = h.QuickCheck == "ok" && h.ForeignKeys
	return h
}
