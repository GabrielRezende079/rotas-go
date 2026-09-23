// Package services implementa os serviços da API: um por função
// (rotas, alternativas, rotas salvas, bases, veículos, info).
package services

import (
	"errors"
	"net/http"
)

// ErroRequisicao carrega o código HTTP e a mensagem de um erro de negócio.
type ErroRequisicao struct {
	Codigo   int
	Mensagem string
}

func (e ErroRequisicao) Error() string { return e.Mensagem }

// codigoDeErro extrai o código HTTP de um erro; erros inesperados viram 500.
func codigoDeErro(err error) int {
	var httpErr ErroRequisicao
	if errors.As(err, &httpErr) {
		return httpErr.Codigo
	}
	return http.StatusInternalServerError
}

// erroHTTP constrói um ErroRequisicao com código e mensagem.
func erroHTTP(codigo int, mensagem string) ErroRequisicao {
	return ErroRequisicao{Codigo: codigo, Mensagem: mensagem}
}

// semPersistência responde 503 quando DATABASE_URL não está configurada.
func semPersistencia() ErroRequisicao {
	return erroHTTP(http.StatusServiceUnavailable, "persistência não configurada (defina DATABASE_URL)")
}
