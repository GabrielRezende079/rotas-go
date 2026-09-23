package services

import (
	"errors"
	"net/http"
	"testing"

	"rotas-go/internal/api/dtos"
)

func TestCalcularRotaRetornaGeometriaEValoresReaisDoGrafo(t *testing.T) {
	g := grafoDeTeste()
	service := NewRouteService(g)

	resp, err := service.CalcularRota(dtos.RotaRequest{
		Origem:    &dtos.Coordenadas{Lat: -20.30, Lng: -40.30},
		Destino:   &dtos.Coordenadas{Lat: -20.32, Lng: -40.32},
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
	service := NewRouteService(g)

	resp, err := service.CalcularRota(dtos.RotaRequest{
		Origem:    &dtos.Coordenadas{Lat: -20.30, Lng: -40.30},
		Destino:   &dtos.Coordenadas{Lat: -20.33, Lng: -40.33},
		Waypoints: []dtos.Coordenadas{{Lat: -20.31, Lng: -40.31}, {Lat: -20.32, Lng: -40.32}},
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
	service := NewRouteService(grafoDeTeste())
	_, err := service.CalcularRota(dtos.RotaRequest{
		Origem:    &dtos.Coordenadas{Lat: -20.30, Lng: -40.30},
		Destino:   &dtos.Coordenadas{Lat: -20.32, Lng: -40.32},
		Algoritmo: "bellman-ford",
	})
	var httpErr ErroRequisicao
	if !errors.As(err, &httpErr) || httpErr.Codigo != http.StatusBadRequest {
		t.Fatalf("erro esperado 400, obtido %v", err)
	}
}

func TestCalcularLoteResolveTodosOsVeiculos(t *testing.T) {
	g := grafoDeTeste()
	service := NewRouteService(g)

	resp, err := service.CalcularLote(dtos.LoteRequest{
		Algoritmo: "dijkstra",
		Veiculos: []dtos.VeiculoRequisicao{
			{ID: "v1", Origem: &dtos.Coordenadas{Lat: -20.30, Lng: -40.30}, Destino: &dtos.Coordenadas{Lat: -20.31, Lng: -40.31}},
			{ID: "v2", Origem: &dtos.Coordenadas{Lat: -20.31, Lng: -40.31}, Destino: &dtos.Coordenadas{Lat: -20.33, Lng: -40.33}},
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
