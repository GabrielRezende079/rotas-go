package routes

import (
	"time"

	"rotas-go/internal/grafo"
)

// Coordenadas é um ponto geográfico, seguindo o contrato da API.
type Coordenadas struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// RotaRequest é o corpo da requisição POST /api/v1/route.
type RotaRequest struct {
	Origem    *Coordenadas `json:"origin" binding:"required"`
	Destino   *Coordenadas `json:"destination" binding:"required"`
	Algoritmo string       `json:"algorithm" binding:"required"`
}

// PontoAjustado representa o ponto real da via usado no cálculo.
type PontoAjustado struct {
	Lat             float64 `json:"lat"`
	Lng             float64 `json:"lng"`
	DistanciaAjuste float64 `json:"snap_distance_km"`
}

// RotaResponse contém a geometria detalhada percorrida no grafo viário.
type RotaResponse struct {
	Origem             PontoAjustado `json:"origin"`
	Destino            PontoAjustado `json:"destination"`
	Algoritmo          string        `json:"algorithm"`
	DistanciaKm        float64       `json:"distance_km"`
	DuracaoEstimadaMin float64       `json:"estimated_duration_minutes"`
	NodosVisitados     int           `json:"nodes_visited"`
	TempoBuscaMs       float64       `json:"search_time_ms"`
	Geometria          []Coordenadas `json:"geometry"`
	Fonte              string        `json:"data_source"`
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
