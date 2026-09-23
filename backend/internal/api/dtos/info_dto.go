package dtos

import "time"

// InfoResponse descreve o conjunto de dados carregado pela API.
type InfoResponse struct {
	Fonte    string    `json:"data_source"`
	GeradoEm time.Time `json:"generated_at"`
	Vertices int       `json:"vertices"`
	Arestas  int64     `json:"edges"`
	ModoReal bool      `json:"real_road_graph"`
}
