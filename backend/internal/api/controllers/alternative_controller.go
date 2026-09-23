package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/api/services"
)

// AlternativeController expõe o handler de rotas alternativas.
type AlternativeController struct {
	service *services.AlternativeService
}

// NewAlternativeController cria o controller de alternativas.
func NewAlternativeController(s *services.AlternativeService) *AlternativeController {
	return &AlternativeController{service: s}
}

// CalcularAlternativas trata POST /api/v1/route/alternatives.
func (ct *AlternativeController) CalcularAlternativas(c *gin.Context) {
	var req dtos.AlternativasRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "corpo da requisição inválido"})
		return
	}
	resp, err := ct.service.CalcularAlternativas(req)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
