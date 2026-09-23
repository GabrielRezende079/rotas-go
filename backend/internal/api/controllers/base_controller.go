package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/api/services"
)

// BaseController expõe os handlers de bases.
type BaseController struct {
	service *services.BaseService
}

// NewBaseController cria o controller de bases.
func NewBaseController(s *services.BaseService) *BaseController {
	return &BaseController{service: s}
}

// CriarBase trata POST /api/v1/bases.
func (ct *BaseController) CriarBase(c *gin.Context) {
	var req dtos.CriarBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "corpo da requisição inválido"})
		return
	}
	resp, err := ct.service.CriarBase(c.Request.Context(), req)
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListarBases trata GET /api/v1/bases?q=&limit=&offset=.
func (ct *BaseController) ListarBases(c *gin.Context) {
	resp, err := ct.service.ListarBases(c.Request.Context(), consultaLista(c))
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ExcluirBase trata DELETE /api/v1/bases/:id.
func (ct *BaseController) ExcluirBase(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	if err := ct.service.ExcluirBase(c.Request.Context(), id); err != nil {
		responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
