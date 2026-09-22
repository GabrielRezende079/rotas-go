package grafo

// Aresta é uma ligação direcionada entre dois nós da malha viária.
type Aresta struct {
	Destino     int64
	DistanciaKm float64
	DuracaoMin  float64
}
