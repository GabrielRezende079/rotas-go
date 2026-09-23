package services

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/grafo"
)

func TestCalcularAlternativasDevolveAlternativasPorCustoTempo(t *testing.T) {
	service := NewAlternativeService(grafoComAlternativas())
	resp, err := service.CalcularAlternativas(dtos.AlternativasRequest{
		Origem:          &dtos.Coordenadas{Lat: -20.0, Lng: -40.0},
		Destino:         &dtos.Coordenadas{Lat: -20.0 - 10.0/1000, Lng: -40.0 - 10.0/1000},
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
	service := NewAlternativeService(g)
	_, err := service.CalcularAlternativas(dtos.AlternativasRequest{
		Origem:  &dtos.Coordenadas{Lat: -20.0, Lng: -40.0},
		Destino: &dtos.Coordenadas{Lat: -20.01, Lng: -40.01},
	})
	var httpErr ErroRequisicao
	if !errors.As(err, &httpErr) || httpErr.Codigo != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422, obtido %v", err)
	}
}

func TestCalcularAlternativasComWaypointsResolveTodosOsTrechos(t *testing.T) {
	service := NewAlternativeService(grafoComAlternativas())
	resp, err := service.CalcularAlternativas(dtos.AlternativasRequest{
		Origem:          &dtos.Coordenadas{Lat: -20.0, Lng: -40.0},
		Destino:         &dtos.Coordenadas{Lat: -20.0 - 10.0/1000, Lng: -40.0 - 10.0/1000},
		Waypoints:       []dtos.Coordenadas{{Lat: -20.0 - 3.0/1000, Lng: -40.0 - 3.0/1000}},
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
