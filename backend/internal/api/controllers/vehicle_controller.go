package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/api/services"
)

// VehicleController expõe os handlers de veículos da frota.
type VehicleController struct {
	service *services.VehicleService
}

// NewVehicleController cria o controller de veículos.
func NewVehicleController(s *services.VehicleService) *VehicleController {
	return &VehicleController{service: s}
}

// CriarVeiculo trata POST /api/v1/vehicles.
func (ct *VehicleController) CriarVeiculo(c *gin.Context) {
	var req dtos.CriarVeiculoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "corpo da requisição inválido"})
		return
	}
	resp, err := ct.service.CriarVeiculo(c.Request.Context(), req)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListarVeiculos trata GET /api/v1/vehicles.
func (ct *VehicleController) ListarVeiculos(c *gin.Context) {
	resp, err := ct.service.ListarVeiculos(c.Request.Context())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, dtos.VeiculosResponse{Veiculos: resp})
}

// ExcluirVeiculo trata DELETE /api/v1/vehicles/:id.
func (ct *VehicleController) ExcluirVeiculo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	if err := ct.service.ExcluirVeiculo(c.Request.Context(), id); err != nil {
		responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
