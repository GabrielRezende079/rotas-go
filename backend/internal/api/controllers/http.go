// Package controllers expõe os handlers HTTP da API.
// Há um controller por função; todos convertem erros de negócio em HTTP.
package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/api/services"
	"rotas-go/internal/storage"
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

// consultaLista extrai os parâmetros q (busca), limit e offset de GETs paginados.
func consultaLista(c *gin.Context) storage.ListaFiltro {
	limite, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	f := storage.ListaFiltro{
		Termo:  strings.TrimSpace(c.Query("q")),
		Limite: limite,
		Offset: offset,
	}
	if f.Limite < 1 {
		f.Limite = 20
	}
	if f.Limite > 50 {
		f.Limite = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return f
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
