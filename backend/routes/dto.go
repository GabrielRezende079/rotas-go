package routes

import (
	"encoding/json"
	"time"

	"rotas-go/internal/grafo"
)

// Coordenadas é um ponto geográfico, seguindo o contrato da API.
type Coordenadas struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// RotaRequest é o corpo da requisição POST /api/v1/route.
// Waypoints são pontos de passagem opcionais entre origem e destino.
// Custo define a métrica minimizada: duracao (padrão) ou distancia.
type RotaRequest struct {
	Origem    *Coordenadas  `json:"origin" binding:"required"`
	Destino   *Coordenadas  `json:"destination" binding:"required"`
	Waypoints []Coordenadas `json:"waypoints"`
	Algoritmo string        `json:"algorithm" binding:"required"`
	Custo     string        `json:"cost"`
}

// PontoAjustado representa o ponto real da via usado no cálculo.
type PontoAjustado struct {
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	DistanciaAjuste float64 `json:"snap_distance_km"`
}

// LegResponse descreve um trecho entre dois pontos consecutivos da rota.
type LegResponse struct {
	Origem             PontoAjustado `json:"origin"`
	Destino            PontoAjustado `json:"destination"`
	DistanciaKm        float64       `json:"distance_km"`
	DuracaoEstimadaMin float64       `json:"estimated_duration_minutes"`
	NodosVisitados     int           `json:"nodes_visited"`
	Geometria          []Coordenadas `json:"geometry"`
}

// RotaResponse contém a geometria detalhada percorrida no grafo viário.
type RotaResponse struct {
	Origem             PontoAjustado   `json:"origin"`
	Destino            PontoAjustado   `json:"destination"`
	Waypoints          []PontoAjustado `json:"waypoints"`
	Algoritmo          string          `json:"algorithm"`
	DistanciaKm        float64         `json:"distance_km"`
	DuracaoEstimadaMin float64         `json:"estimated_duration_minutes"`
	NodosVisitados     int             `json:"nodes_visited"`
	TempoBuscaMs       float64         `json:"search_time_ms"`
	Geometria          []Coordenadas   `json:"geometry"`
	Legs               []LegResponse   `json:"legs"`
	Fonte              string          `json:"data_source"`
}

// VeiculoRequisicao representa a rota de um único veículo dentro de um lote.
type VeiculoRequisicao struct {
	ID        string        `json:"id" binding:"required"`
	Origem    *Coordenadas  `json:"origin" binding:"required"`
	Destino   *Coordenadas  `json:"destination" binding:"required"`
	Waypoints []Coordenadas `json:"waypoints"`
}

// LoteRequest é o corpo da requisição POST /api/v1/routes.
type LoteRequest struct {
	Algoritmo string              `json:"algorithm" binding:"required"`
	Veiculos  []VeiculoRequisicao `json:"vehicles" binding:"required,min=1"`
	Custo     string              `json:"cost"`
}

// LoteItemResponse é o resultado calculado de um veículo dentro do lote.
type LoteItemResponse struct {
	ID string `json:"id"`
	RotaResponse
}

// LoteResponse é a resposta de POST /api/v1/routes.
type LoteResponse struct {
	Rotas []LoteItemResponse `json:"routes"`
}

// SalvarRotaRequest é o corpo da requisição POST /api/v1/routes/saved.
// Rotas contém o snapshot calculado pelo cliente (com geometria).
type SalvarRotaRequest struct {
	Nome      string             `json:"name" binding:"required"`
	Algoritmo string             `json:"algorithm" binding:"required"`
	Rotas     []LoteItemResponse `json:"routes" binding:"required,min=1"`
}

// RotaSalvaSummary é o resumo exibido na listagem de rotas salvas.
type RotaSalvaSummary struct {
	ID          int64     `json:"id"`
	Nome        string    `json:"name"`
	Algoritmo   string    `json:"algorithm"`
	CriadaEm    time.Time `json:"created_at"`
	Veiculos    int       `json:"vehicle_count"`
	DistanciaKm float64   `json:"total_distance_km"`
	DuracaoMin  float64   `json:"total_duration_minutes"`
}

// RotaSalvaDetalhe inclui o snapshot completo da rota salva.
type RotaSalvaDetalhe struct {
	RotaSalvaSummary
	Rotas json.RawMessage `json:"routes"`
}

// AlternativasRequest é o corpo de POST /api/v1/route/alternatives.
type AlternativasRequest struct {
	Origem          *Coordenadas  `json:"origin" binding:"required"`
	Destino         *Coordenadas  `json:"destination" binding:"required"`
	Waypoints       []Coordenadas `json:"waypoints"`
	Custo           string        `json:"cost"`
	MaxAlternativas int           `json:"max_alternatives"`
}

// AlternativaResponse é uma das rotas alternativas de um trecho.
type AlternativaResponse struct {
	DistanciaKm        float64       `json:"distance_km"`
	DuracaoEstimadaMin float64       `json:"estimated_duration_minutes"`
	NodosVisitados     int           `json:"nodes_visited"`
	CustoTotal         float64       `json:"total_cost"`
	Geometria          []Coordenadas `json:"geometry"`
}

// TrechoAlternativas reúne as alternativas de um par origem/destino.
type TrechoAlternativas struct {
	Origem       PontoAjustado         `json:"origin"`
	Destino      PontoAjustado         `json:"destination"`
	Alternativas []AlternativaResponse `json:"alternatives"`
}

// AlternativasResponse é a resposta de POST /api/v1/route/alternatives.
type AlternativasResponse struct {
	Custo   string               `json:"cost"`
	Trechos []TrechoAlternativas `json:"legs"`
}

// InfoResponse descreve o conjunto de dados carregado pela API.
type InfoResponse struct {
	Fonte    string    `json:"data_source"`
	GeradoEm time.Time `json:"generated_at"`
	Vertices int       `json:"vertices"`
	Arestas  int64     `json:"edges"`
	ModoReal bool      `json:"real_road_graph"`
}

func infoDoGrafo(meta grafo.Metadata) InfoResponse {
	return InfoResponse{
		Fonte:    meta.Fonte,
		GeradoEm: meta.GeradoEm,
		Vertices: meta.Vertices,
		Arestas:  meta.Arestas,
		ModoReal: meta.Fonte != "grafo didático embutido",
	}
}
