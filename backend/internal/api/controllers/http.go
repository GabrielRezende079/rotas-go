// Package controllers expõe os handlers HTTP da API.
// Há um controller por função; todos convertem erros de negócio em HTTP.
package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/api/services"
)

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

// responderErro converte os erros de negócio em respostas HTTP.
func responderErro(c *gin.Context, err error) {
	var httpErr services.ErroRequisicao
	if errors.As(err, &httpErr) {
		c.JSON(httpErr.Codigo, gin.H{"error": httpErr.Mensagem})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "erro interno"})
}
