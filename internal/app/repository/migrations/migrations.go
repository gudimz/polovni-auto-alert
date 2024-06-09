package migrations

import "embed"

//go:embed files/*.sql
var Files embed.FS
