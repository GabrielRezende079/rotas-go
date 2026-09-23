// Package dtos define os tipos de troca com a API HTTP.
// Os tipos aqui não dependem do domínio interno; o contrato JSON é fixo.
package dtos

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
