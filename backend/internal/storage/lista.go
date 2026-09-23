// Package storage centraliza a persistência em PostgreSQL.
package storage

// ListaFiltro parametriza listagens paginadas com busca textual.
// Termo é buscado com ILIKE "%termo%"; quando vazio, nada é filtrado.
// Limite e Offset controlam a paginação (Limite 0 é normalizado pelo serviço).
type ListaFiltro struct {
	Termo  string
	Limite int
	Offset int
}
