package routes

import (
	"testing"
	"time"

	"rotas-go/internal/grafo"
)

func TestCalcularRotaRetornaGeometriaEValoresReaisDoGrafo(t *testing.T) {
	g := grafo.NovoGrafo()
	g.AddVertice(grafo.Vertice{ID: 1, Lat: -20.30, Lng: -40.30})
	g.AddVertice(grafo.Vertice{ID: 2, Lat: -20.31, Lng: -40.31})
	g.AddVertice(grafo.Vertice{ID: 3, Lat: -20.32, Lng: -40.32})
	g.AddArestaBidirecional(1, 2, 1.5, 2)
	g.AddArestaBidirecional(2, 3, 1.7, 3)
	g.SetMetadata("OSM teste", time.Unix(1, 0))
	service := NewRouteService(g)

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
	if resp.Fonte != "OSM teste" {
		t.Fatalf("fonte = %q", resp.Fonte)
	}
}
