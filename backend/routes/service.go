package routes

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"rotas-go/internal/grafo"
)

const distanciaMaximaAjusteKm = 25.0

type erroRequisicao struct {
	codigo   int
	mensagem string
}

func (e erroRequisicao) Error() string { return e.mensagem }

// RouteService orquestra o grafo e o algoritmo de busca escolhido.
type RouteService struct {
	grafo *grafo.Grafo
}

func NewRouteService(g *grafo.Grafo) *RouteService { return &RouteService{grafo: g} }

// Info informa qual malha está efetivamente carregada.
func (s *RouteService) Info() InfoResponse { return infoDoGrafo(s.grafo.Metadata()) }

// CalcularRota ajusta os pontos à via mais próxima e executa o algoritmo pedido.
func (s *RouteService) CalcularRota(req RotaRequest) (RotaResponse, error) {
	if req.Origem == nil || req.Destino == nil {
		return RotaResponse{}, erroRequisicao{http.StatusBadRequest, "origin e destination são obrigatórios"}
	}
	if req.Origem.Lat < -90 || req.Origem.Lat > 90 || req.Destino.Lat < -90 || req.Destino.Lat > 90 {
		return RotaResponse{}, erroRequisicao{http.StatusBadRequest, "latitude fora do intervalo [-90, 90]"}
	}
	if req.Origem.Lng < -180 || req.Origem.Lng > 180 || req.Destino.Lng < -180 || req.Destino.Lng > 180 {
		return RotaResponse{}, erroRequisicao{http.StatusBadRequest, "longitude fora do intervalo [-180, 180]"}
	}

	algoritmo := strings.ToLower(strings.TrimSpace(req.Algoritmo))
	if algoritmo != "dijkstra" && algoritmo != "astar" {
		return RotaResponse{}, erroRequisicao{
			http.StatusBadRequest,
			fmt.Sprintf("algoritmo desconhecido: %s", req.Algoritmo),
		}
	}

	origem, ajusteOrigem, ok := s.grafo.VerticeMaisProximo(req.Origem.Lat, req.Origem.Lng)
	if !ok {
		return RotaResponse{}, erroRequisicao{http.StatusServiceUnavailable, "grafo sem vértices"}
	}
	destino, ajusteDestino, ok := s.grafo.VerticeMaisProximo(req.Destino.Lat, req.Destino.Lng)
	if !ok {
		return RotaResponse{}, erroRequisicao{http.StatusServiceUnavailable, "grafo sem vértices"}
	}
	if ajusteOrigem > distanciaMaximaAjusteKm || ajusteDestino > distanciaMaximaAjusteKm {
		return RotaResponse{}, erroRequisicao{
			http.StatusUnprocessableEntity,
			"um dos pontos está distante demais da malha viária do Espírito Santo",
		}
	}

	inicioBusca := time.Now()
	var resultado grafo.ResultadoBusca
	if algoritmo == "dijkstra" {
		resultado = s.grafo.Dijkstra(origem.ID, destino.ID)
	} else {
		resultado = s.grafo.AEstrela(origem.ID, destino.ID)
	}
	tempoBuscaMs := float64(time.Since(inicioBusca).Microseconds()) / 1000
	if len(resultado.Caminho) == 0 {
		return RotaResponse{}, erroRequisicao{
			http.StatusUnprocessableEntity,
			"não existe caminho dirigível entre os pontos selecionados",
		}
	}

	geometria := make([]Coordenadas, len(resultado.Caminho))
	for i, v := range resultado.Caminho {
		geometria[i] = Coordenadas{Lat: v.Lat, Lng: v.Lng}
	}
	meta := s.grafo.Metadata()
	return RotaResponse{
		Origem: PontoAjustado{
			Lat: origem.Lat, Lng: origem.Lng, DistanciaAjuste: arredondar(ajusteOrigem, 3),
		},
		Destino: PontoAjustado{
			Lat: destino.Lat, Lng: destino.Lng, DistanciaAjuste: arredondar(ajusteDestino, 3),
		},
		Algoritmo:          algoritmo,
		DistanciaKm:        arredondar(resultado.DistanciaKm, 2),
		DuracaoEstimadaMin: arredondar(resultado.DuracaoMin, 1),
		NodosVisitados:     resultado.NodosVisitados,
		TempoBuscaMs:       arredondar(tempoBuscaMs, 3),
		Geometria:          geometria,
		Fonte:              meta.Fonte,
	}, nil
}

func arredondar(valor float64, casas int) float64 {
	fator := math.Pow(10, float64(casas))
	return math.Round(valor*fator) / fator
}
