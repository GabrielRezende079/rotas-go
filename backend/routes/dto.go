package routes

// Coordenadas é um ponto geográfico, seguindo o contrato da API.
type Coordenadas struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// RotaRequest é o corpo da requisição POST /api/v1/route.
type RotaRequest struct {
	Origem    Coordenadas `json:"origin"`
	Destino   Coordenadas `json:"destination"`
	Algoritmo string      `json:"algorithm"`
}

// NoDTO representa um vértice na resposta da API.
type NoDTO struct {
	Nome string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// RotaResponse é a resposta de uma rota calculada.
type RotaResponse struct {
	Origem         NoDTO   `json:"origin"`
	Destino        NoDTO   `json:"destination"`
	Algoritmo      string  `json:"algorithm"`
	DistanciaKm    float64 `json:"distance_km"`
	NodosVisitados int     `json:"nodes_visited"`
	Caminho        []NoDTO `json:"path"`
}
