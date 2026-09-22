package grafo

import (
	"math"
	"time"

	geo "rotas-go/internal/math"
)

// Metadata identifica a origem e a data de geração do grafo.
type Metadata struct {
	Fonte    string    `json:"source"`
	GeradoEm time.Time `json:"generated_at"`
	Vertices int       `json:"vertices"`
	Arestas  int64     `json:"edges"`
}

// Grafo armazena a malha viária como lista de adjacências direcionadas.
type Grafo struct {
	vertices    map[int64]Vertice
	adjacencias map[int64][]Aresta
	numArestas  int64
	metadata    Metadata
}

// ResultadoBusca é o retorno comum de Dijkstra e A*.
type ResultadoBusca struct {
	Caminho        []Vertice
	Arestas        []Aresta
	DistanciaKm    float64
	DuracaoMin     float64
	NodosVisitados int
}

// NovoGrafo cria um grafo vazio.
func NovoGrafo() *Grafo {
	return &Grafo{
		vertices:    make(map[int64]Vertice),
		adjacencias: make(map[int64][]Aresta),
	}
}

// AddVertice adiciona ou atualiza um vértice.
func (g *Grafo) AddVertice(v Vertice) {
	g.vertices[v.ID] = v
}

// AddArestaDirecionada cria uma ligação em apenas um sentido.
func (g *Grafo) AddArestaDirecionada(origem, destino int64, distanciaKm, duracaoMin float64) {
	g.adjacencias[origem] = append(g.adjacencias[origem], Aresta{
		Destino:     destino,
		DistanciaKm: distanciaKm,
		DuracaoMin:  duracaoMin,
	})
	g.numArestas++
}

// AddArestaBidirecional cria uma ligação com o mesmo custo nos dois sentidos.
func (g *Grafo) AddArestaBidirecional(origem, destino int64, distanciaKm, duracaoMin float64) {
	g.AddArestaDirecionada(origem, destino, distanciaKm, duracaoMin)
	g.AddArestaDirecionada(destino, origem, distanciaKm, duracaoMin)
}

// Vertices devolve uma cópia dos vértices. Deve ser usada apenas em grafos pequenos.
func (g *Grafo) Vertices() []Vertice {
	lista := make([]Vertice, 0, len(g.vertices))
	for _, v := range g.vertices {
		lista = append(lista, v)
	}
	return lista
}

// QuantidadeVertices informa o tamanho do grafo.
func (g *Grafo) QuantidadeVertices() int { return len(g.vertices) }

// QuantidadeArestas informa a quantidade de ligações direcionadas.
func (g *Grafo) QuantidadeArestas() int64 { return g.numArestas }

// Metadata devolve informações sobre a geração do grafo.
func (g *Grafo) Metadata() Metadata {
	m := g.metadata
	m.Vertices = len(g.vertices)
	m.Arestas = g.numArestas
	return m
}

// SetMetadata registra a procedência dos dados.
func (g *Grafo) SetMetadata(fonte string, geradoEm time.Time) {
	g.metadata.Fonte = fonte
	g.metadata.GeradoEm = geradoEm
}

// VerticeMaisProximo devolve o nó viário mais próximo e a distância do encaixe.
func (g *Grafo) VerticeMaisProximo(lat, lng float64) (Vertice, float64, bool) {
	var melhor Vertice
	menor := math.MaxFloat64
	encontrado := false
	for _, v := range g.vertices {
		d := geo.HaversineKm(lat, lng, v.Lat, v.Lng)
		if d < menor {
			menor = d
			melhor = v
			encontrado = true
		}
	}
	return melhor, menor, encontrado
}

// NovoGrafoESDidatico preserva o grafo reduzido da demonstração inicial.
// Ele é somente um fallback; as rotas reais usam o grafo importado do OSM.
func NovoGrafoESDidatico() *Grafo {
	g := NovoGrafo()

	type municipio struct {
		id       int64
		lat, lng float64
	}
	municipios := map[string]municipio{
		"Vitória":               {-1, -20.3155, -40.3128},
		"Vila Velha":            {-2, -20.3297, -40.2925},
		"Serra":                 {-3, -20.1288, -40.3078},
		"Cariacica":             {-4, -20.2635, -40.4166},
		"Viana":                 {-5, -20.3928, -40.4961},
		"Domingos Martins":      {-6, -20.3633, -40.6592},
		"Colatina":              {-7, -19.5390, -40.6300},
		"Santa Leopoldina":      {-8, -20.0990, -40.5290},
		"Santa Maria de Jetibá": {-9, -20.0250, -40.6960},
		"Fundão":                {-10, -19.9300, -40.4070},
		"Aracruz":               {-11, -19.8200, -40.2760},
		"Ibiraçu":               {-12, -19.8332, -40.3699},
		"João Neiva":            {-13, -19.7570, -40.3860},
		"Linhares":              {-14, -19.3920, -40.0720},
		"São Mateus":            {-15, -18.7160, -39.8590},
		"Guarapari":             {-16, -20.6572, -40.5109},
		"Anchieta":              {-17, -20.8055, -40.6425},
		"Piúma":                 {-18, -20.8350, -40.7290},
	}
	for _, p := range municipios {
		g.AddVertice(Vertice{ID: p.id, Lat: p.lat, Lng: p.lng})
	}

	add := func(a, b string, km float64) {
		// Velocidade didática de 60 km/h: duração em minutos igual à distância.
		g.AddArestaBidirecional(municipios[a].id, municipios[b].id, km, km)
	}
	add("Vitória", "Vila Velha", 11)
	add("Vitória", "Serra", 16)
	add("Vitória", "Cariacica", 8)
	add("Vila Velha", "Cariacica", 12)
	add("Cariacica", "Viana", 10)
	add("Viana", "Domingos Martins", 22)
	add("Domingos Martins", "Colatina", 75)
	add("Vitória", "Santa Leopoldina", 30)
	add("Cariacica", "Santa Leopoldina", 20)
	add("Santa Leopoldina", "Santa Maria de Jetibá", 25)
	add("Santa Leopoldina", "Fundão", 28)
	add("Serra", "Fundão", 13)
	add("Fundão", "Aracruz", 20)
	add("Aracruz", "Ibiraçu", 14)
	add("Ibiraçu", "João Neiva", 5)
	add("João Neiva", "Linhares", 36)
	add("Linhares", "São Mateus", 70)
	add("Linhares", "Colatina", 63)
	add("Serra", "Santa Maria de Jetibá", 42)
	add("Viana", "Guarapari", 32)
	add("Vila Velha", "Guarapari", 40)
	add("Guarapari", "Anchieta", 24)
	add("Anchieta", "Piúma", 9)
	g.SetMetadata("grafo didático embutido", time.Now().UTC())
	return g
}
