package grafo

import (
	"math"

	geo "rotas-go/internal/math"
)

// Grafo armazena o mapa como lista de adjacências.
type Grafo struct {
	vertices    map[string]*Vertice
	adjacencias map[string][]Aresta
}

// ResultadoBusca é o retorno comum de Dijkstra e A*.
type ResultadoBusca struct {
	Caminho        []*Vertice
	DistanciaKm    float64
	NodosVisitados int
}

// NovoGrafo cria um grafo vazio.
func NovoGrafo() *Grafo {
	return &Grafo{
		vertices:    make(map[string]*Vertice),
		adjacencias: make(map[string][]Aresta),
	}
}

// AddVertice adiciona um vértice ao grafo.
func (g *Grafo) AddVertice(v Vertice) {
	g.vertices[v.ID] = &v
}

// AddAresta conecta dois vértices nos dois sentidos (grafo não direcionado).
func (g *Grafo) AddAresta(origem, destino string, pesoKm float64) {
	g.adjacencias[origem] = append(g.adjacencias[origem], Aresta{Origem: origem, Destino: destino, PesoKm: pesoKm})
	g.adjacencias[destino] = append(g.adjacencias[destino], Aresta{Origem: destino, Destino: origem, PesoKm: pesoKm})
}

// Vertices devolve todos os vértices do grafo.
func (g *Grafo) Vertices() []*Vertice {
	lista := make([]*Vertice, 0, len(g.vertices))
	for _, v := range g.vertices {
		lista = append(lista, v)
	}
	return lista
}

// VerticeMaisProximo devolve o vértice mais próximo de um ponto (lat, lng).
func (g *Grafo) VerticeMaisProximo(lat, lng float64) *Vertice {
	var melhor *Vertice
	menor := math.MaxFloat64
	for _, v := range g.vertices {
		d := geo.HaversineKm(lat, lng, v.Lat, v.Lng)
		if d < menor {
			menor = d
			melhor = v
		}
	}
	return melhor
}

// NovoGrafoES monta o grafo com os municípios do Espírito Santo e suas rodovias.
func NovoGrafoES() *Grafo {
	g := NovoGrafo()

	municipios := map[string]struct {
		lat float64
		lng float64
	}{
		"Vitória":               {-20.3155, -40.3128},
		"Vila Velha":            {-20.3297, -40.2925},
		"Serra":                 {-20.1288, -40.3078},
		"Cariacica":             {-20.2635, -40.4166},
		"Viana":                 {-20.3928, -40.4961},
		"Domingos Martins":      {-20.3633, -40.6592},
		"Colatina":              {-19.5390, -40.6300},
		"Santa Leopoldina":      {-20.0990, -40.5290},
		"Santa Maria de Jetibá": {-20.0250, -40.6960},
		"Fundão":                {-19.9300, -40.4070},
		"Aracruz":               {-19.8200, -40.2760},
		"Ibiraçu":               {-19.8332, -40.3699},
		"João Neiva":            {-19.7570, -40.3860},
		"Linhares":              {-19.3920, -40.0720},
		"São Mateus":            {-18.7160, -39.8590},
		"Guarapari":             {-20.6572, -40.5109},
		"Anchieta":              {-20.8055, -40.6425},
		"Piúma":                 {-20.8350, -40.7290},
	}
	for nome, p := range municipios {
		g.AddVertice(Vertice{ID: nome, Nome: nome, Lat: p.lat, Lng: p.lng})
	}

	// Rodovias entre os municípios (peso em km aproximado por estrada real).
	g.AddAresta("Vitória", "Vila Velha", 11) // Terceira Ponte
	g.AddAresta("Vitória", "Serra", 16)      // BR-101
	g.AddAresta("Vitória", "Cariacica", 8)   // BR-262
	g.AddAresta("Vila Velha", "Cariacica", 12)
	g.AddAresta("Cariacica", "Viana", 10)        // BR-262
	g.AddAresta("Viana", "Domingos Martins", 22) // BR-262
	g.AddAresta("Domingos Martins", "Colatina", 75)
	g.AddAresta("Vitória", "Santa Leopoldina", 30)               // ES-080
	g.AddAresta("Cariacica", "Santa Leopoldina", 20)             // ES-161
	g.AddAresta("Santa Leopoldina", "Santa Maria de Jetibá", 25) // ES-080
	g.AddAresta("Santa Leopoldina", "Fundão", 28)                // ES-080
	g.AddAresta("Serra", "Fundão", 13)                           // BR-101
	g.AddAresta("Fundão", "Aracruz", 20)                         // BR-101
	g.AddAresta("Aracruz", "Ibiraçu", 14)                        // BR-101
	g.AddAresta("Ibiraçu", "João Neiva", 5)                      // BR-101
	g.AddAresta("João Neiva", "Linhares", 36)                    // BR-101
	g.AddAresta("Linhares", "São Mateus", 70)                    // BR-101
	g.AddAresta("Linhares", "Colatina", 63)                      // BR-259
	g.AddAresta("Serra", "Santa Maria de Jetibá", 42)
	g.AddAresta("Viana", "Guarapari", 32)      // ES-060
	g.AddAresta("Vila Velha", "Guarapari", 40) // ES-060 Rod. do Sol
	g.AddAresta("Guarapari", "Anchieta", 24)   // ES-060
	g.AddAresta("Anchieta", "Piúma", 9)        // ES-060

	return g
}
