package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/api/services"
)

// RouteController expõe os handlers de cálculo de rotas e lotes.
type RouteController struct {
	service *services.RouteService
}

// NewRouteController cria o controller de rotas.
func NewRouteController(s *services.RouteService) *RouteController {
	return &RouteController{service: s}
}

// CalcularRota trata POST /api/v1/route.
func (ct *RouteController) CalcularRota(c *gin.Context) {
	var req dtos.RotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "corpo da requisição inválido"})
		return
	}
	resp, err := ct.service.CalcularRota(req)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CalcularLote trata POST /api/v1/routes.
func (ct *RouteController) CalcularLote(c *gin.Context) {
	var req dtos.LoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "corpo da requisição inválido"})
		return
	}
	resp, err := ct.service.CalcularLote(req)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
