package services

import "rotas-go/internal/storage"

// normalizarFiltro garante valores válidos para a paginação:
// Limite padrão 20, máximo 50; Offset mínimo 0.
func normalizarFiltro(f storage.ListaFiltro) storage.ListaFiltro {
	if f.Limite <= 0 {
		f.Limite = 20
	}
	if f.Limite > 50 {
		f.Limite = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return f
}
