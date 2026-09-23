package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RouteController expõe os handlers HTTP da API.
type RouteController struct {
	service *RouteService
}

// NewRouteController cria o controller com o serviço de rotas.
func NewRouteController(s *RouteService) *RouteController {
	return &RouteController{service: s}
}

// CORSMiddleware libera o acesso de qualquer origem, método e cabeçalho.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, Origin")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Info trata GET /api/v1/info.
func (ct *RouteController) Info(c *gin.Context) {
	c.JSON(http.StatusOK, ct.service.Info())
}

// CalcularRota trata POST /api/v1/route.
func (ct *RouteController) CalcularRota(c *gin.Context) {
	var req RotaRequest
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
	var req LoteRequest
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

// CalcularAlternativas trata POST /api/v1/route/alternatives.
func (ct *RouteController) CalcularAlternativas(c *gin.Context) {
	var req AlternativasRequest
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

// SalvarRota trata POST /api/v1/routes/saved.
func (ct *RouteController) SalvarRota(c *gin.Context) {
	var req SalvarRotaRequest
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

// ListarRotasSalvas trata GET /api/v1/routes/saved.
func (ct *RouteController) ListarRotasSalvas(c *gin.Context) {
	resp, err := ct.service.ListarRotasSalvas(c.Request.Context())
	if err != nil {
		responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// BuscarRotaSalva trata GET /api/v1/routes/saved/:id.
func (ct *RouteController) BuscarRotaSalva(c *gin.Context) {
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
func (ct *RouteController) ExcluirRotaSalva(c *gin.Context) {
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

func responderErro(c *gin.Context, err error) {
	var httpErr erroRequisicao
	if errors.As(err, &httpErr) {
		c.JSON(httpErr.codigo, gin.H{"error": httpErr.mensagem})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno"})
}
