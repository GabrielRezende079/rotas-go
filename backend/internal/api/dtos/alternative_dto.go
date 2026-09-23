package dtos

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
