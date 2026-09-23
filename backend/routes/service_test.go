package routes

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"rotas-go/internal/grafo"
	"rotas-go/internal/storage"
)

func grafoDeTeste() *grafo.Grafo {
	g := grafo.NovoGrafo()
	g.AddVertice(grafo.Vertice{ID: 1, Lat: -20.30, Lng: -40.30})
	g.AddVertice(grafo.Vertice{ID: 2, Lat: -20.31, Lng: -40.31})
	g.AddVertice(grafo.Vertice{ID: 3, Lat: -20.32, Lng: -40.32})
	g.AddVertice(grafo.Vertice{ID: 4, Lat: -20.33, Lng: -40.33})
	g.AddArestaBidirecional(1, 2, 1.5, 2)
	g.AddArestaBidirecional(2, 3, 1.7, 3)
	g.AddArestaBidirecional(3, 4, 0.8, 1)
	g.SetMetadata("OSM teste", time.Unix(1, 0))
	return g
}

func TestCalcularRotaRetornaGeometriaEValoresReaisDoGrafo(t *testing.T) {
	g := grafoDeTeste()
	service := NewRouteService(g, nil)

	resp, err := service.CalcularRota(RotaRequest{
		Origem:    &Coordenadas{Lat: -20.30, Lng: -40.30},
		Destino:   &Coordenadas{Lat: -20.32, Lng: -40.32},
		Algoritmo: "dijkstra",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.DistanciaKm != 3.2 || resp.DuracaoEstimadaMin != 5 {
		t.Fatalf("métricas inesperadas: distância=%f duração=%f", resp.DistanciaKm, resp.DuracaoEstimadaMin)
	}
	if len(resp.Geometria) != 3 {
		t.Fatalf("geometria tem %d pontos, esperado 3", len(resp.Geometria))
	}
	if len(resp.Legs) != 1 {
		t.Fatalf("legs tem %d trechos, esperado 1", len(resp.Legs))
	}
	if resp.Fonte != "OSM teste" {
		t.Fatalf("fonte = %q", resp.Fonte)
	}
}

func TestCalcularRotaComWaypointsEncadeiaTrechos(t *testing.T) {
	g := grafoDeTeste()
	service := NewRouteService(g, nil)

	resp, err := service.CalcularRota(RotaRequest{
		Origem:    &Coordenadas{Lat: -20.30, Lng: -40.30},
		Destino:   &Coordenadas{Lat: -20.33, Lng: -40.33},
		Waypoints: []Coordenadas{{Lat: -20.31, Lng: -40.31}, {Lat: -20.32, Lng: -40.32}},
		Algoritmo: "astar",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.DistanciaKm != 4.0 {
		t.Fatalf("distância = %f, esperado 4.0", resp.DistanciaKm)
	}
	if resp.DuracaoEstimadaMin != 6 {
		t.Fatalf("duração = %f, esperado 6", resp.DuracaoEstimadaMin)
	}
	if len(resp.Legs) != 3 {
		t.Fatalf("legs tem %d trechos, esperado 3", len(resp.Legs))
	}
	if len(resp.Waypoints) != 2 {
		t.Fatalf("waypoints tem %d pontos, esperado 2", len(resp.Waypoints))
	}
	if len(resp.Geometria) != 4 {
		t.Fatalf("geometria tem %d pontos (vértices duplicados não eliminados), esperado 4", len(resp.Geometria))
	}
}

func TestCalcularRotaRejeitaAlgoritmoDesconhecido(t *testing.T) {
	service := NewRouteService(grafoDeTeste(), nil)
	_, err := service.CalcularRota(RotaRequest{
		Origem:    &Coordenadas{Lat: -20.30, Lng: -40.30},
		Destino:   &Coordenadas{Lat: -20.32, Lng: -40.32},
		Algoritmo: "bellman-ford",
	})
	var httpErr erroRequisicao
	if !errors.As(err, &httpErr) || httpErr.codigo != http.StatusBadRequest {
		t.Fatalf("erro esperado 400, obtido %v", err)
	}
}

func TestCalcularLoteResolveTodosOsVeiculos(t *testing.T) {
	g := grafoDeTeste()
	service := NewRouteService(g, nil)

	resp, err := service.CalcularLote(LoteRequest{
		Algoritmo: "dijkstra",
		Veiculos: []VeiculoRequisicao{
			{ID: "v1", Origem: &Coordenadas{Lat: -20.30, Lng: -40.30}, Destino: &Coordenadas{Lat: -20.31, Lng: -40.31}},
			{ID: "v2", Origem: &Coordenadas{Lat: -20.31, Lng: -40.31}, Destino: &Coordenadas{Lat: -20.33, Lng: -40.33}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Rotas) != 2 {
		t.Fatalf("lote devolveu %d rotas, esperado 2", len(resp.Rotas))
	}
	if resp.Rotas[0].ID != "v1" || resp.Rotas[1].ID != "v2" {
		t.Fatalf("ordem dos veículos não preservada")
	}
	if resp.Rotas[0].DistanciaKm != 1.5 || resp.Rotas[1].DistanciaKm != 2.5 {
		t.Fatalf("métricas por veículo incorretas: %f %f", resp.Rotas[0].DistanciaKm, resp.Rotas[1].DistanciaKm)
	}
}

func TestSalvarRotaSemPersistenciaRetorna503(t *testing.T) {
	service := NewRouteService(grafoDeTeste(), nil)
	_, err := service.SalvarRota(context.Background(), SalvarRotaRequest{
		Nome: "teste", Algoritmo: "dijkstra",
	})
	var httpErr erroRequisicao
	if err == nil || !errors.As(err, &httpErr) || httpErr.codigo != http.StatusServiceUnavailable {
		t.Fatalf("esperava 503 sem persistência, obtido %v", err)
	}
}

func TestSalvarRotaUsaStoreFornecido(t *testing.T) {
	store := &storeFake{}
	service := NewRouteService(grafoDeTeste(), store)
	resumo, err := service.SalvarRota(context.Background(), SalvarRotaRequest{
		Nome:      "entrega norte",
		Algoritmo: "astar",
		Rotas: []LoteItemResponse{
			{ID: "v1", RotaResponse: RotaResponse{DistanciaKm: 10, DuracaoEstimadaMin: 15}},
			{ID: "v2", RotaResponse: RotaResponse{DistanciaKm: 20, DuracaoEstimadaMin: 25}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resumo.Veiculos != 2 || resumo.DistanciaKm != 30 || resumo.DuracaoMin != 40 {
		t.Fatalf("resumo incorreto: %+v", resumo)
	}
	if store.ultimoNome != "entrega norte" || store.ultimoAlgoritmo != "astar" {
		t.Fatalf("store chamado com valores errados: %q %q", store.ultimoNome, store.ultimoAlgoritmo)
	}
}

// storeFake simula a persistência sem depender de um banco real.
type storeFake struct {
	ultimoNome      string
	ultimoAlgoritmo string
}

func (f *storeFake) Salvar(_ context.Context, nome, algoritmo string, veiculos int, distanciaKm, duracaoMin float64, _ []byte) (storage.RotaSalva, error) {
	f.ultimoNome = nome
	f.ultimoAlgoritmo = algoritmo
	return storage.RotaSalva{
		ID: 7, Nome: nome, Algoritmo: algoritmo,
		Veiculos: veiculos, DistanciaKm: distanciaKm, DuracaoMin: duracaoMin,
	}, nil
}

func (f *storeFake) Listar(_ context.Context) ([]storage.RotaSalva, error) {
	return nil, nil
}

func (f *storeFake) Buscar(_ context.Context, _ int64) (storage.RotaSalva, error) {
	return storage.RotaSalva{}, nil
}

func (f *storeFake) Excluir(_ context.Context, _ int64) error { return nil }

func grafoComAlternativas() *grafo.Grafo {
	g := grafo.NovoGrafo()
	for id := 1; id <= 10; id++ {
		g.AddVertice(grafo.Vertice{ID: int64(id), Lat: -20.0 - float64(id)/1000, Lng: -40.0 - float64(id)/1000})
	}
	// Corredor 1 (mais rápido): 1-2-3-4-5-6-10
	for i := 1; i <= 5; i++ {
		g.AddArestaBidirecional(int64(i), int64(i+1), 1, 2)
	}
	g.AddArestaBidirecional(6, 10, 1, 2)
	// Corredor 2 (mais lento): 1-7-8-9-10
	for _, par := range [][2]int64{{1, 7}, {7, 8}, {8, 9}, {9, 10}} {
		g.AddArestaBidirecional(par[0], par[1], 10, 4)
	}
	g.SetMetadata("OSM teste", time.Unix(1, 0))
	return g
}

func TestCalcularAlternativasDevolveAlternativasPorCustoTempo(t *testing.T) {
	service := NewRouteService(grafoComAlternativas(), nil)
	resp, err := service.CalcularAlternativas(AlternativasRequest{
		Origem:          &Coordenadas{Lat: -20.0, Lng: -40.0},
		Destino:         &Coordenadas{Lat: -20.0 - 10.0/1000, Lng: -40.0 - 10.0/1000},
		MaxAlternativas: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Trechos) != 1 {
		t.Fatalf("trechos = %d, esperado 1", len(resp.Trechos))
	}
	trecho := resp.Trechos[0]
	if len(trecho.Alternativas) < 2 {
		t.Fatalf("alternativas = %d, esperado ao menos 2", len(trecho.Alternativas))
	}
	primeira, segunda := trecho.Alternativas[0], trecho.Alternativas[1]
	if primeira.CustoTotal > segunda.CustoTotal {
		t.Fatalf("alternativas fora de ordem de custo: %f > %f", primeira.CustoTotal, segunda.CustoTotal)
	}
	if len(primeira.Geometria) == 0 || len(segunda.Geometria) == 0 {
		t.Fatalf("geometria vazia em alternativa")
	}
	if resp.Custo != "duration" {
		t.Fatalf("custo = %q, esperado duration", resp.Custo)
	}
}

func TestCalcularAlternativasSemCaminhoRetorna422(t *testing.T) {
	g := grafo.NovoGrafo()
	g.AddVertice(grafo.Vertice{ID: 1, Lat: -20.0, Lng: -40.0})
	g.AddVertice(grafo.Vertice{ID: 2, Lat: -20.01, Lng: -40.01})
	g.SetMetadata("OSM teste", time.Unix(1, 0))
	service := NewRouteService(g, nil)
	_, err := service.CalcularAlternativas(AlternativasRequest{
		Origem:  &Coordenadas{Lat: -20.0, Lng: -40.0},
		Destino: &Coordenadas{Lat: -20.01, Lng: -40.01},
	})
	var httpErr erroRequisicao
	if !errors.As(err, &httpErr) || httpErr.codigo != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422, obtido %v", err)
	}
}

func TestCalcularAlternativasComWaypointsResolveTodosOsTrechos(t *testing.T) {
	service := NewRouteService(grafoComAlternativas(), nil)
	resp, err := service.CalcularAlternativas(AlternativasRequest{
		Origem:          &Coordenadas{Lat: -20.0, Lng: -40.0},
		Destino:         &Coordenadas{Lat: -20.0 - 10.0/1000, Lng: -40.0 - 10.0/1000},
		Waypoints:       []Coordenadas{{Lat: -20.0 - 3.0/1000, Lng: -40.0 - 3.0/1000}},
		MaxAlternativas: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Trechos) != 2 {
		t.Fatalf("trechos = %d, esperado 2", len(resp.Trechos))
	}
	for _, trecho := range resp.Trechos {
		if len(trecho.Alternativas) == 0 {
			t.Fatalf("trecho sem alternativas: %+v", trecho)
		}
	}
}
