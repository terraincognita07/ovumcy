package migrations

import "embed"

// Files stores forward-only SQL migrations embedded into the binary.
//
//go:embed *.sql
var Files embed.FS
