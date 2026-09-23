package dtos

import "time"

// BaseResponse representa uma base persistida (localização padrão).
type BaseResponse struct {
	ID       int64     `json:"id"`
	Nome     string    `json:"name"`
	Lat      float64   `json:"lat"`
	Lng      float64   `json:"lng"`
	CriadaEm time.Time `json:"created_at"`
}

// CriarBaseRequest é o corpo de POST /api/v1/bases.
type CriarBaseRequest struct {
	Nome string  `json:"name" binding:"required"`
	Lat  float64 `json:"lat" binding:"required"`
	Lng  float64 `json:"lng" binding:"required"`
}

// BasesResponse é a resposta de GET /api/v1/bases.
type BasesResponse struct {
	Bases []BaseResponse `json:"bases"`
}
