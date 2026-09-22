package routes

import (
	"errors"
	"net/http"

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
		var httpErr erroRequisicao
		if errors.As(err, &httpErr) {
			c.JSON(httpErr.codigo, gin.H{"error": httpErr.mensagem})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
