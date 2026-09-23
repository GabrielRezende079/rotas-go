package dtos

import "time"

// VeiculoResponse representa um veículo da frota cadastrado.
type VeiculoResponse struct {
	ID           int64     `json:"id"`
	Modelo       string    `json:"modelo"`
	Categoria    string    `json:"categoria"`
	Placa        string    `json:"placa"`
	Status       string    `json:"status"`
	Kilometragem float64   `json:"kilometragem"`
	Velocidade   float64   `json:"velocidade"`
	CriadoEm     time.Time `json:"created_at"`
}

// CriarVeiculoRequest é o corpo de POST /api/v1/vehicles.
type CriarVeiculoRequest struct {
	Modelo       string  `json:"modelo" binding:"required"`
	Categoria    string  `json:"categoria" binding:"required"`
	Placa        string  `json:"placa" binding:"required"`
	Status       string  `json:"status" binding:"required"`
	Kilometragem float64 `json:"kilometragem"`
	Velocidade   float64 `json:"velocidade"`
}

// VeiculosResponse é a resposta de GET /api/v1/vehicles.
type VeiculosResponse struct {
	Veiculos []VeiculoResponse `json:"vehicles"`
}
