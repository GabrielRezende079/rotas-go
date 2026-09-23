package services

import (
	"rotas-go/internal/api/dtos"
	"rotas-go/internal/grafo"
)

// InfoService descreve a malha viária efetivamente carregada.
type InfoService struct {
	grafo *grafo.Grafo
}

// NewInfoService cria o serviço de informação do grafo.
func NewInfoService(g *grafo.Grafo) *InfoService {
	return &InfoService{grafo: g}
}

// Info devolve os metadados do grafo carregado.
func (s *InfoService) Info() dtos.InfoResponse {
	meta := s.grafo.Metadata()
	return dtos.InfoResponse{
		Fonte:    meta.Fonte,
		GeradoEm: meta.GeradoEm,
		Vertices: meta.Vertices,
		Arestas:  meta.Arestas,
		ModoReal: meta.Fonte != "grafo didático embutido",
	}
}
