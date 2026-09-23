package services

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/grafo"
)

const (
	distanciaMaximaAjusteKm = 25.0
	maxVeiculosPorLote      = 50
)

// RouteService calcula rotas sobre o grafo viário (Dijkstra/A*).
// A persistência vive em SavedRouteService/BaseService/VehicleService.
type RouteService struct {
	grafo *grafo.Grafo
}

// NewRouteService cria o serviço de cálculo de rotas.
func NewRouteService(g *grafo.Grafo) *RouteService {
	return &RouteService{grafo: g}
}

// CalcularRota calcula uma rota entre origem e destino, passando pelos
// waypoints (pontos de passagem) informados.
func (s *RouteService) CalcularRota(req dtos.RotaRequest) (dtos.RotaResponse, error) {
	if req.Origem == nil || req.Destino == nil {
		return dtos.RotaResponse{}, erroHTTP(http.StatusBadRequest, "origin e destination são obrigatórios")
	}
	pontos := make([]dtos.Coordenadas, 0, 2+len(req.Waypoints))
	pontos = append(pontos, *req.Origem)
	pontos = append(pontos, req.Waypoints...)
	pontos = append(pontos, *req.Destino)
	return s.calcularRotaOrdenada(pontos, req.Algoritmo, req.Custo)
}

// CalcularLote calcula a rota de vários veículos em uma única requisição.
// Cada veículo é resolvido em paralelo; o grafo é somente leitura durante
// a busca, então as execuções concorrentes são seguras.
func (s *RouteService) CalcularLote(req dtos.LoteRequest) (dtos.LoteResponse, error) {
	if len(req.Veiculos) == 0 {
		return dtos.LoteResponse{}, erroHTTP(http.StatusBadRequest, "informe ao menos um veículo")
	}
	if len(req.Veiculos) > maxVeiculosPorLote {
		return dtos.LoteResponse{}, erroHTTP(
			http.StatusBadRequest,
			fmt.Sprintf("limite de veículos por lote é %d", maxVeiculosPorLote),
		)
	}
	algoritmo := strings.ToLower(strings.TrimSpace(req.Algoritmo))
	if !algoritmoValido(algoritmo) {
		return dtos.LoteResponse{}, erroHTTP(http.StatusBadRequest, fmt.Sprintf("algoritmo desconhecido: %s", req.Algoritmo))
	}
	custo := normalizarCusto(req.Custo)

	if len(req.Veiculos) == 1 {
		rota, err := s.calcularRotaVeiculo(req.Veiculos[0], algoritmo, custo)
		if err != nil {
			return dtos.LoteResponse{}, err
		}
		return dtos.LoteResponse{Rotas: []dtos.LoteItemResponse{rota}}, nil
	}

	canal := make(chan resultadoVeiculo, len(req.Veiculos))
	for i, v := range req.Veiculos {
		go func(indice int, veiculo dtos.VeiculoRequisicao) {
			rota, err := s.calcularRotaVeiculo(veiculo, algoritmo, custo)
			canal <- resultadoVeiculo{indice: indice, rota: rota, err: err}
		}(i, v)
	}

	respostas := make([]dtos.LoteItemResponse, len(req.Veiculos))
	var primeiroErro error
	for range req.Veiculos {
		r := <-canal
		if r.err != nil {
			if primeiroErro == nil {
				primeiroErro = r.err
			}
			continue
		}
		respostas[r.indice] = r.rota
	}
	if primeiroErro != nil {
		return dtos.LoteResponse{}, primeiroErro
	}
	return dtos.LoteResponse{Rotas: respostas}, nil
}

type resultadoVeiculo struct {
	indice int
	rota   dtos.LoteItemResponse
	err    error
}

func (s *RouteService) calcularRotaVeiculo(v dtos.VeiculoRequisicao, algoritmo, custo string) (dtos.LoteItemResponse, error) {
	if v.ID == "" {
		return dtos.LoteItemResponse{}, erroHTTP(http.StatusBadRequest, "veículo sem identificador")
	}
	if v.Origem == nil || v.Destino == nil {
		return dtos.LoteItemResponse{}, erroHTTP(
			http.StatusBadRequest,
			fmt.Sprintf("veículo %q exige origin e destination", v.ID),
		)
	}
	pontos := make([]dtos.Coordenadas, 0, 2+len(v.Waypoints))
	pontos = append(pontos, *v.Origem)
	pontos = append(pontos, v.Waypoints...)
	pontos = append(pontos, *v.Destino)
	rota, err := s.calcularRotaOrdenada(pontos, algoritmo, custo)
	if err != nil {
		return dtos.LoteItemResponse{}, erroHTTP(codigoDeErro(err), fmt.Sprintf("veículo %q: %s", v.ID, err.Error()))
	}
	return dtos.LoteItemResponse{ID: v.ID, RotaResponse: rota}, nil
}

// calcularRotaOrdenada encadeia a busca pelos pares consecutivos de pontos,
// agrega métricas e concatena a geometria sem duplicar vértices compartilhados.
func (s *RouteService) calcularRotaOrdenada(pontos []dtos.Coordenadas, algoritmo, custo string) (dtos.RotaResponse, error) {
	if len(pontos) < 2 {
		return dtos.RotaResponse{}, erroHTTP(http.StatusBadRequest, "são necessários ao menos dois pontos para traçar uma rota")
	}
	algoritmo = strings.ToLower(strings.TrimSpace(algoritmo))
	if !algoritmoValido(algoritmo) {
		return dtos.RotaResponse{}, erroHTTP(http.StatusBadRequest, fmt.Sprintf("algoritmo desconhecido: %s", algoritmo))
	}
	custo = normalizarCusto(custo)

	vertices := make([]grafo.Vertice, 0, len(pontos))
	ajustados := make([]dtos.PontoAjustado, 0, len(pontos))
	for _, p := range pontos {
		v, ajuste, err := ajustarPonto(s.grafo, p)
		if err != nil {
			return dtos.RotaResponse{}, err
		}
		vertices = append(vertices, v)
		ajustados = append(ajustados, ajuste)
	}

	inicio := time.Now()
	legs := make([]dtos.LegResponse, 0, len(pontos)-1)
	geometria := make([]dtos.Coordenadas, 0)
	var totalDistancia, totalDuracao float64
	var totalVisitados int

	for i := 0; i < len(pontos)-1; i++ {
		resultado := s.buscar(vertices[i].ID, vertices[i+1].ID, algoritmo, custo)
		if len(resultado.Caminho) == 0 {
			return dtos.RotaResponse{}, erroHTTP(
				http.StatusUnprocessableEntity,
				"não existe caminho dirigível entre os pontos selecionados",
			)
		}

		legCoords := make([]dtos.Coordenadas, len(resultado.Caminho))
		for j, vert := range resultado.Caminho {
			legCoords[j] = dtos.Coordenadas{Lat: vert.Lat, Lng: vert.Lng}
		}
		legs = append(legs, dtos.LegResponse{
			Origem:             ajustados[i],
			Destino:            ajustados[i+1],
			DistanciaKm:        arredondar(resultado.DistanciaKm, 2),
			DuracaoEstimadaMin: arredondar(resultado.DuracaoMin, 1),
			NodosVisitados:     resultado.NodosVisitados,
			Geometria:          legCoords,
		})
		if i == 0 {
			geometria = append(geometria, legCoords...)
		} else {
			geometria = append(geometria, legCoords[1:]...)
		}
		totalDistancia += resultado.DistanciaKm
		totalDuracao += resultado.DuracaoMin
		totalVisitados += resultado.NodosVisitados
	}
	tempoBuscaMs := float64(time.Since(inicio).Microseconds()) / 1000

	meta := s.grafo.Metadata()
	return dtos.RotaResponse{
		Origem:             ajustados[0],
		Destino:            ajustados[len(ajustados)-1],
		Waypoints:          ajustados[1 : len(ajustados)-1],
		Algoritmo:          algoritmo,
		DistanciaKm:        arredondar(totalDistancia, 2),
		DuracaoEstimadaMin: arredondar(totalDuracao, 1),
		NodosVisitados:     totalVisitados,
		TempoBuscaMs:       arredondar(tempoBuscaMs, 3),
		Geometria:          geometria,
		Legs:               legs,
		Fonte:              meta.Fonte,
	}, nil
}

func (s *RouteService) buscar(origemID, destinoID int64, algoritmo, custo string) grafo.ResultadoBusca {
	if algoritmo == "dijkstra" {
		if custo == "distance" {
			return s.grafo.Dijkstra(origemID, destinoID)
		}
		return s.grafo.DijkstraPorDuracao(origemID, destinoID)
	}
	if custo == "distance" {
		return s.grafo.AEstrela(origemID, destinoID)
	}
	return s.grafo.AEstrelaPorDuracao(origemID, destinoID)
}

// ajustarPonto valida a coordenada e a amarra ao vértice mais próximo da malha.
func ajustarPonto(g *grafo.Grafo, c dtos.Coordenadas) (grafo.Vertice, dtos.PontoAjustado, error) {
	if c.Lat < -90 || c.Lat > 90 {
		return grafo.Vertice{}, dtos.PontoAjustado{}, erroHTTP(http.StatusBadRequest, "latitude fora do intervalo [-90, 90]")
	}
	if c.Lng < -180 || c.Lng > 180 {
		return grafo.Vertice{}, dtos.PontoAjustado{}, erroHTTP(http.StatusBadRequest, "longitude fora do intervalo [-180, 180]")
	}
	v, ajuste, ok := g.VerticeMaisProximo(c.Lat, c.Lng)
	if !ok {
		return grafo.Vertice{}, dtos.PontoAjustado{}, erroHTTP(http.StatusServiceUnavailable, "grafo sem vértices")
	}
	if ajuste > distanciaMaximaAjusteKm {
		return grafo.Vertice{}, dtos.PontoAjustado{}, erroHTTP(
			http.StatusUnprocessableEntity,
			"um dos pontos está distante demais da malha viária do Espírito Santo",
		)
	}
	return v, dtos.PontoAjustado{Lat: v.Lat, Lng: v.Lng, DistanciaAjuste: arredondar(ajuste, 3)}, nil
}

// normalizarCusto aceita "distance" ou vazio; qualquer outro valor cai para
// "duration", que é a métrica padrão das alternativas.
func normalizarCusto(c string) string {
	if strings.ToLower(strings.TrimSpace(c)) == "distance" {
		return "distance"
	}
	return "duration"
}

func algoritmoValido(a string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	return a == "dijkstra" || a == "astar"
}

func arredondar(valor float64, casas int) float64 {
	fator := math.Pow(10, float64(casas))
	return math.Round(valor*fator) / fator
}
