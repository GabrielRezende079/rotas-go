package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/api/services"
)

// SavedRouteController expõe os handlers de rotas salvas.
type SavedRouteController struct {
	service *services.SavedRouteService
}

// NewSavedRouteController cria o controller de rotas salvas.
func NewSavedRouteController(s *services.SavedRouteService) *SavedRouteController {
	return &SavedRouteController{service: s}
}

// SalvarRota trata POST /api/v1/routes/saved.
func (ct *SavedRouteController) SalvarRota(c *gin.Context) {
	var req dtos.SalvarRotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "corpo da requisição inválido"})
		return
	}
	resp, err := ct.service.SalvarRota(c.Request.Context(), req)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListarRotasSalvas trata GET /api/v1/routes/saved?q=&limit=&offset=.
func (ct *SavedRouteController) ListarRotasSalvas(c *gin.Context) {
	resp, err := ct.service.ListarRotasSalvas(c.Request.Context(), consultaLista(c))
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// BuscarRotaSalva trata GET /api/v1/routes/saved/:id.
func (ct *SavedRouteController) BuscarRotaSalva(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	resp, err := ct.service.BuscarRotaSalva(c.Request.Context(), id)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ExcluirRotaSalva trata DELETE /api/v1/routes/saved/:id.
func (ct *SavedRouteController) ExcluirRotaSalva(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	if err := ct.service.ExcluirRotaSalva(c.Request.Context(), id); err != nil {
		responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
