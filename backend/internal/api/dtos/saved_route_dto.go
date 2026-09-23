package dtos

import (
	"encoding/json"
	"time"
)

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
