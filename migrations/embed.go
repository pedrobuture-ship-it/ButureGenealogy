// Package migrations embeds the versioned SQL migrations for farm.db.
// Files are named NNNN_description.sql and applied in order, each in a transaction.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
