package services

import (
	"net/http"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/grafo"
)

// AlternativeService calcula rotas alternativas (Yen + dissimilaridade)
// para um par origem/destino, com ou sem waypoints.
type AlternativeService struct {
	grafo *grafo.Grafo
}

// NewAlternativeService cria o serviço de rotas alternativas.
func NewAlternativeService(g *grafo.Grafo) *AlternativeService {
	return &AlternativeService{grafo: g}
}

// CalcularAlternativas devolve, para cada trecho origem/destino, até três rotas
// alternativas mais rápidas. Os trechos são resolvidos em paralelo.
func (s *AlternativeService) CalcularAlternativas(req dtos.AlternativasRequest) (dtos.AlternativasResponse, error) {
	if req.Origem == nil || req.Destino == nil {
		return dtos.AlternativasResponse{}, erroHTTP(http.StatusBadRequest, "origin e destination são obrigatórios")
	}
	custo := normalizarCusto(req.Custo)
	if len(req.Waypoints) == 0 {
		pontos := []dtos.Coordenadas{*req.Origem, *req.Destino}
		trecho, err := s.alternativasDeTrecho(pontos, custo, req.MaxAlternativas)
		if err != nil {
			return dtos.AlternativasResponse{}, err
		}
		return dtos.AlternativasResponse{Custo: custo, Trechos: []dtos.TrechoAlternativas{trecho}}, nil
	}

	pontos := make([]dtos.Coordenadas, 0, 2+len(req.Waypoints))
	pontos = append(pontos, *req.Origem)
	pontos = append(pontos, req.Waypoints...)
	pontos = append(pontos, *req.Destino)

	canal := make(chan resultadoTrecho, len(pontos)-1)
	for i := 0; i < len(pontos)-1; i++ {
		go func(inicio int, fim int) {
			trecho, err := s.alternativasDeTrecho(pontos[inicio:fim+1], custo, req.MaxAlternativas)
			canal <- resultadoTrecho{indice: inicio, trecho: trecho, err: err}
		}(i, i+1)
	}
	respostas := make([]dtos.TrechoAlternativas, len(pontos)-1)
	var primeiroErro error
	for range pontos[:len(pontos)-1] {
		r := <-canal
		if r.err != nil {
			if primeiroErro == nil {
				primeiroErro = r.err
			}
			continue
		}
		respostas[r.indice] = r.trecho
	}
	if primeiroErro != nil {
		return dtos.AlternativasResponse{}, primeiroErro
	}
	return dtos.AlternativasResponse{Custo: custo, Trechos: respostas}, nil
}

type resultadoTrecho struct {
	indice int
	trecho dtos.TrechoAlternativas
	err    error
}

func (s *AlternativeService) alternativasDeTrecho(pontos []dtos.Coordenadas, custo string, maxAlternativas int) (dtos.TrechoAlternativas, error) {
	vertices := make([]grafo.Vertice, len(pontos))
	ajustados := make([]dtos.PontoAjustado, len(pontos))
	for i, p := range pontos {
		v, ajuste, err := ajustarPonto(s.grafo, p)
		if err != nil {
			return dtos.TrechoAlternativas{}, err
		}
		vertices[i] = v
		ajustados[i] = ajuste
	}
	resultados := s.grafo.KMenoresCaminhos(vertices[0].ID, vertices[len(vertices)-1].ID, maxAlternativas)
	if len(resultados) == 0 {
		return dtos.TrechoAlternativas{}, erroHTTP(
			http.StatusUnprocessableEntity,
			"não existe caminho dirigível entre os pontos selecionados",
		)
	}
	alternativas := make([]dtos.AlternativaResponse, 0, len(resultados))
	for _, resultado := range resultados {
		geometria := make([]dtos.Coordenadas, len(resultado.Caminho))
		for j, vert := range resultado.Caminho {
			geometria[j] = dtos.Coordenadas{Lat: vert.Lat, Lng: vert.Lng}
		}
		custoTotal := resultado.DuracaoMin
		if custo == "distance" {
			custoTotal = resultado.DistanciaKm
		}
		alternativas = append(alternativas, dtos.AlternativaResponse{
			DistanciaKm:        arredondar(resultado.DistanciaKm, 2),
			DuracaoEstimadaMin: arredondar(resultado.DuracaoMin, 1),
			NodosVisitados:     resultado.NodosVisitados,
			CustoTotal:         arredondar(custoTotal, 2),
			Geometria:          geometria,
		})
	}
	return dtos.TrechoAlternativas{
		Origem:       ajustados[0],
		Destino:      ajustados[len(ajustados)-1],
		Alternativas: alternativas,
	}, nil
}
