package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/api/services"
)

// InfoController responde sobre o grafo carregado.
type InfoController struct {
	service *services.InfoService
}

// NewInfoController cria o controller de informações do grafo.
func NewInfoController(s *services.InfoService) *InfoController {
	return &InfoController{service: s}
}

// Info trata GET /api/v1/info.
func (ct *InfoController) Info(c *gin.Context) {
	c.JSON(http.StatusOK, ct.service.Info())
}
