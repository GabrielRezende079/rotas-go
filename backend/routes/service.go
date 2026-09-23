package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"rotas-go/internal/grafo"
	"rotas-go/internal/storage"
)

const (
	distanciaMaximaAjusteKm = 25.0
	maxVeiculosPorLote      = 50
)

type erroRequisicao struct {
	codigo   int
	mensagem string
}

func (e erroRequisicao) Error() string { return e.mensagem }

// SavedRouteStore é a persistência usada para guardar rotas calculadas.
// O RouteService depende apenas desta interface; a implementação PostgreSQL
// vive em rotas-go/internal/storage.
type SavedRouteStore interface {
	Salvar(ctx context.Context, nome, algoritmo string, veiculos int, distanciaKm, duracaoMin float64, dados []byte) (storage.RotaSalva, error)
	Listar(ctx context.Context) ([]storage.RotaSalva, error)
	Buscar(ctx context.Context, id int64) (storage.RotaSalva, error)
	Excluir(ctx context.Context, id int64) error
}

// RouteService orquestra o grafo, o algoritmo de busca e a persistência.
type RouteService struct {
	grafo *grafo.Grafo
	store SavedRouteStore
}

func NewRouteService(g *grafo.Grafo, store SavedRouteStore) *RouteService {
	return &RouteService{grafo: g, store: store}
}

// Info informa qual malha está efetivamente carregada.
func (s *RouteService) Info() InfoResponse { return infoDoGrafo(s.grafo.Metadata()) }

// CalcularRota calcula uma rota entre origem e destino, passando pelos
// waypoints (pontos de passagem) informados.
func (s *RouteService) CalcularRota(req RotaRequest) (RotaResponse, error) {
	if req.Origem == nil || req.Destino == nil {
		return RotaResponse{}, erroRequisicao{http.StatusBadRequest, "origin e destination são obrigatórios"}
	}
	pontos := make([]Coordenadas, 0, 2+len(req.Waypoints))
	pontos = append(pontos, *req.Origem)
	pontos = append(pontos, req.Waypoints...)
	pontos = append(pontos, *req.Destino)
	return s.calcularRotaOrdenada(pontos, req.Algoritmo)
}

// CalcularLote calcula a rota de vários veículos em uma única requisição.
// Cada veículo é resolvido em paralelo; o grafo é somente leitura durante
// a busca, então as execuções concorrentes são seguras.
func (s *RouteService) CalcularLote(req LoteRequest) (LoteResponse, error) {
	if len(req.Veiculos) == 0 {
		return LoteResponse{}, erroRequisicao{http.StatusBadRequest, "informe ao menos um veículo"}
	}
	if len(req.Veiculos) > maxVeiculosPorLote {
		return LoteResponse{}, erroRequisicao{
			http.StatusBadRequest,
			fmt.Sprintf("limite de veículos por lote é %d", maxVeiculosPorLote),
		}
	}
	algoritmo := strings.ToLower(strings.TrimSpace(req.Algoritmo))
	if !algoritmoValido(algoritmo) {
		return LoteResponse{}, erroRequisicao{http.StatusBadRequest, fmt.Sprintf("algoritmo desconhecido: %s", req.Algoritmo)}
	}

	if len(req.Veiculos) == 1 {
		rota, err := s.calcularRotaVeiculo(req.Veiculos[0], algoritmo)
		if err != nil {
			return LoteResponse{}, err
		}
		return LoteResponse{Rotas: []LoteItemResponse{rota}}, nil
	}

	canal := make(chan resultadoVeiculo, len(req.Veiculos))
	for i, v := range req.Veiculos {
		go func(indice int, veiculo VeiculoRequisicao) {
			rota, err := s.calcularRotaVeiculo(veiculo, algoritmo)
			canal <- resultadoVeiculo{indice: indice, rota: rota, err: err}
		}(i, v)
	}

	respostas := make([]LoteItemResponse, len(req.Veiculos))
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
		return LoteResponse{}, primeiroErro
	}
	return LoteResponse{Rotas: respostas}, nil
}

type resultadoVeiculo struct {
	indice int
	rota   LoteItemResponse
	err    error
}

func (s *RouteService) calcularRotaVeiculo(v VeiculoRequisicao, algoritmo string) (LoteItemResponse, error) {
	if v.ID == "" {
		return LoteItemResponse{}, erroRequisicao{http.StatusBadRequest, "veículo sem identificador"}
	}
	if v.Origem == nil || v.Destino == nil {
		return LoteItemResponse{}, erroRequisicao{
			http.StatusBadRequest,
			fmt.Sprintf("veículo %q exige origin e destination", v.ID),
		}
	}
	pontos := make([]Coordenadas, 0, 2+len(v.Waypoints))
	pontos = append(pontos, *v.Origem)
	pontos = append(pontos, v.Waypoints...)
	pontos = append(pontos, *v.Destino)
	rota, err := s.calcularRotaOrdenada(pontos, algoritmo)
	if err != nil {
		return LoteItemResponse{}, erroRequisicao{codigoDeErro(err), fmt.Sprintf("veículo %q: %s", v.ID, err.Error())}
	}
	return LoteItemResponse{ID: v.ID, RotaResponse: rota}, nil
}

// calcularRotaOrdenada encadeia a busca pelos pares consecutivos de pontos,
// agrega métricas e concatena a geometria sem duplicar vértices compartilhados.
func (s *RouteService) calcularRotaOrdenada(pontos []Coordenadas, algoritmo string) (RotaResponse, error) {
	if len(pontos) < 2 {
		return RotaResponse{}, erroRequisicao{http.StatusBadRequest, "são necessários ao menos dois pontos para traçar uma rota"}
	}
	algoritmo = strings.ToLower(strings.TrimSpace(algoritmo))
	if !algoritmoValido(algoritmo) {
		return RotaResponse{}, erroRequisicao{http.StatusBadRequest, fmt.Sprintf("algoritmo desconhecido: %s", algoritmo)}
	}

	vertices := make([]grafo.Vertice, 0, len(pontos))
	ajustados := make([]PontoAjustado, 0, len(pontos))
	for _, p := range pontos {
		v, ajuste, err := s.ajustarPonto(p)
		if err != nil {
			return RotaResponse{}, err
		}
		vertices = append(vertices, v)
		ajustados = append(ajustados, ajuste)
	}

	inicio := time.Now()
	legs := make([]LegResponse, 0, len(pontos)-1)
	geometria := make([]Coordenadas, 0)
	var totalDistancia, totalDuracao float64
	var totalVisitados int

	for i := 0; i < len(pontos)-1; i++ {
		var resultado grafo.ResultadoBusca
		if algoritmo == "dijkstra" {
			resultado = s.grafo.Dijkstra(vertices[i].ID, vertices[i+1].ID)
		} else {
			resultado = s.grafo.AEstrela(vertices[i].ID, vertices[i+1].ID)
		}
		if len(resultado.Caminho) == 0 {
			return RotaResponse{}, erroRequisicao{
				http.StatusUnprocessableEntity,
				"não existe caminho dirigível entre os pontos selecionados",
			}
		}

		legCoords := make([]Coordenadas, len(resultado.Caminho))
		for j, vert := range resultado.Caminho {
			legCoords[j] = Coordenadas{Lat: vert.Lat, Lng: vert.Lng}
		}
		legs = append(legs, LegResponse{
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
	return RotaResponse{
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

func (s *RouteService) ajustarPonto(c Coordenadas) (grafo.Vertice, PontoAjustado, error) {
	if c.Lat < -90 || c.Lat > 90 {
		return grafo.Vertice{}, PontoAjustado{}, erroRequisicao{http.StatusBadRequest, "latitude fora do intervalo [-90, 90]"}
	}
	if c.Lng < -180 || c.Lng > 180 {
		return grafo.Vertice{}, PontoAjustado{}, erroRequisicao{http.StatusBadRequest, "longitude fora do intervalo [-180, 180]"}
	}
	v, ajuste, ok := s.grafo.VerticeMaisProximo(c.Lat, c.Lng)
	if !ok {
		return grafo.Vertice{}, PontoAjustado{}, erroRequisicao{http.StatusServiceUnavailable, "grafo sem vértices"}
	}
	if ajuste > distanciaMaximaAjusteKm {
		return grafo.Vertice{}, PontoAjustado{}, erroRequisicao{
			http.StatusUnprocessableEntity,
			"um dos pontos está distante demais da malha viária do Espírito Santo",
		}
	}
	return v, PontoAjustado{Lat: v.Lat, Lng: v.Lng, DistanciaAjuste: arredondar(ajuste, 3)}, nil
}

// SalvarRota persiste o snapshot calculado de um lote de rotas.
func (s *RouteService) SalvarRota(ctx context.Context, req SalvarRotaRequest) (RotaSalvaSummary, error) {
	if s.store == nil {
		return RotaSalvaSummary{}, erroRequisicao{http.StatusServiceUnavailable, "persistência não configurada (defina DATABASE_URL)"}
	}
	nome := strings.TrimSpace(req.Nome)
	if nome == "" {
		return RotaSalvaSummary{}, erroRequisicao{http.StatusBadRequest, "informe um nome para a rota"}
	}
	algoritmo := strings.ToLower(strings.TrimSpace(req.Algoritmo))
	if !algoritmoValido(algoritmo) {
		return RotaSalvaSummary{}, erroRequisicao{http.StatusBadRequest, fmt.Sprintf("algoritmo desconhecido: %s", req.Algoritmo)}
	}
	if len(req.Rotas) == 0 {
		return RotaSalvaSummary{}, erroRequisicao{http.StatusBadRequest, "a rota não possui veículos"}
	}

	var totalDistancia, totalDuracao float64
	for _, r := range req.Rotas {
		totalDistancia += r.DistanciaKm
		totalDuracao += r.DuracaoEstimadaMin
	}
	dados, err := json.Marshal(req.Rotas)
	if err != nil {
		return RotaSalvaSummary{}, err
	}
	salva, err := s.store.Salvar(ctx, nome, algoritmo, len(req.Rotas), totalDistancia, totalDuracao, dados)
	if err != nil {
		return RotaSalvaSummary{}, err
	}
	return resumoDeRotaSalva(salva), nil
}

// ListarRotasSalvas devolve o resumo de todas as rotas salvas.
func (s *RouteService) ListarRotasSalvas(ctx context.Context) ([]RotaSalvaSummary, error) {
	if s.store == nil {
		return nil, erroRequisicao{http.StatusServiceUnavailable, "persistência não configurada (defina DATABASE_URL)"}
	}
	lista, err := s.store.Listar(ctx)
	if err != nil {
		return nil, err
	}
	resumo := make([]RotaSalvaSummary, 0, len(lista))
	for _, r := range lista {
		resumo = append(resumo, resumoDeRotaSalva(r))
	}
	return resumo, nil
}

// BuscarRotaSalva devolve o snapshot completo de uma rota salva.
func (s *RouteService) BuscarRotaSalva(ctx context.Context, id int64) (RotaSalvaDetalhe, error) {
	if s.store == nil {
		return RotaSalvaDetalhe{}, erroRequisicao{http.StatusServiceUnavailable, "persistência não configurada (defina DATABASE_URL)"}
	}
	salva, err := s.store.Buscar(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNaoEncontrada) {
			return RotaSalvaDetalhe{}, erroRequisicao{http.StatusNotFound, "rota salva não encontrada"}
		}
		return RotaSalvaDetalhe{}, err
	}
	return RotaSalvaDetalhe{RotaSalvaSummary: resumoDeRotaSalva(salva), Rotas: salva.Dados}, nil
}

// ExcluirRotaSalva remove uma rota salva.
func (s *RouteService) ExcluirRotaSalva(ctx context.Context, id int64) error {
	if s.store == nil {
		return erroRequisicao{http.StatusServiceUnavailable, "persistência não configurada (defina DATABASE_URL)"}
	}
	err := s.store.Excluir(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNaoEncontrada) {
			return erroRequisicao{http.StatusNotFound, "rota salva não encontrada"}
		}
		return err
	}
	return nil
}

func resumoDeRotaSalva(r storage.RotaSalva) RotaSalvaSummary {
	return RotaSalvaSummary{
		ID:          r.ID,
		Nome:        r.Nome,
		Algoritmo:   r.Algoritmo,
		CriadaEm:    r.CriadaEm,
		Veiculos:    r.Veiculos,
		DistanciaKm: r.DistanciaKm,
		DuracaoMin:  r.DuracaoMin,
	}
}

func algoritmoValido(a string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	return a == "dijkstra" || a == "astar"
}

func codigoDeErro(err error) int {
	var httpErr erroRequisicao
	if errors.As(err, &httpErr) {
		return httpErr.codigo
	}
	return http.StatusInternalServerError
}

func arredondar(valor float64, casas int) float64 {
	fator := math.Pow(10, float64(casas))
	return math.Round(valor*fator) / fator
}
