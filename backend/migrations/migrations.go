// Package migrations empacota os arquivos SQL aplicados no startup da API.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
