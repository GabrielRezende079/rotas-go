package routes

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"

	"rotas-go/internal/grafo"
)

// erroRequisicao carrega o código HTTP adequado para o erro.
type erroRequisicao struct {
	codigo   int
	mensagem string
}

func (e erroRequisicao) Error() string { return e.mensagem }

// RouteService orquestra o grafo e o algoritmo de busca escolhido.
type RouteService struct {
	grafo *grafo.Grafo
}

// NewRouteService cria o serviço a partir do grafo montado.
func NewRouteService(g *grafo.Grafo) *RouteService {
	return &RouteService{grafo: g}
}

// ListarNos devolve todos os municípios do grafo, em ordem alfabética.
func (s *RouteService) ListarNos() []NoDTO {
	vertices := s.grafo.Vertices()
	sort.Slice(vertices, func(i, j int) bool { return vertices[i].Nome < vertices[j].Nome })
	nos := make([]NoDTO, 0, len(vertices))
	for _, v := range vertices {
		nos = append(nos, NoDTO{Nome: v.Nome, Lat: v.Lat, Lng: v.Lng})
	}
	return nos
}

// CalcularRota ajusta origem/destino aos vértices mais próximos (snap) e
// aplica o algoritmo pedido, retornando o caminho e métricas de estudo.
func (s *RouteService) CalcularRota(req RotaRequest) (RotaResponse, error) {
	if req.Origem.Lat < -90 || req.Origem.Lat > 90 || req.Destino.Lat < -90 || req.Destino.Lat > 90 {
		return RotaResponse{}, erroRequisicao{http.StatusBadRequest, "latitude fora do intervalo [-90, 90]"}
	}
	if req.Origem.Lng < -180 || req.Origem.Lng > 180 || req.Destino.Lng < -180 || req.Destino.Lng > 180 {
		return RotaResponse{}, erroRequisicao{http.StatusBadRequest, "longitude fora do intervalo [-180, 180]"}
	}

	algoritmo := strings.ToLower(strings.TrimSpace(req.Algoritmo))
	if algoritmo == "" {
		return RotaResponse{}, erroRequisicao{http.StatusBadRequest, "o campo algorithm é obrigatório"}
	}
	if algoritmo != "dijkstra" && algoritmo != "astar" {
		return RotaResponse{}, erroRequisicao{
			http.StatusNotFound,
			fmt.Sprintf("algoritmo desconhecido: %s", req.Algoritmo),
		}
	}

	origem := s.grafo.VerticeMaisProximo(req.Origem.Lat, req.Origem.Lng)
	destino := s.grafo.VerticeMaisProximo(req.Destino.Lat, req.Destino.Lng)

	var resultado grafo.ResultadoBusca
	if algoritmo == "dijkstra" {
		resultado = s.grafo.Dijkstra(origem.ID, destino.ID)
	} else {
		resultado = s.grafo.AEstrela(origem.ID, destino.ID)
	}

	if len(resultado.Caminho) == 0 {
		return RotaResponse{}, erroRequisicao{http.StatusInternalServerError, "caminho não encontrado entre os pontos"}
	}

	caminho := make([]NoDTO, 0, len(resultado.Caminho))
	for _, v := range resultado.Caminho {
		caminho = append(caminho, NoDTO{Nome: v.Nome, Lat: v.Lat, Lng: v.Lng})
	}

	return RotaResponse{
		Origem:         caminho[0],
		Destino:        caminho[len(caminho)-1],
		Algoritmo:      algoritmo,
		DistanciaKm:    math.Round(resultado.DistanciaKm*10) / 10,
		NodosVisitados: resultado.NodosVisitados,
		Caminho:        caminho,
	}, nil
}
