package grafo

import (
	"math"
	"path/filepath"
	"testing"
	"time"
)

func grafoDeTeste() *Grafo {
	g := NovoGrafo()
	g.AddVertice(Vertice{ID: 1, Lat: -20.30, Lng: -40.30})
	g.AddVertice(Vertice{ID: 2, Lat: -20.31, Lng: -40.31})
	g.AddVertice(Vertice{ID: 3, Lat: -20.32, Lng: -40.32})
	g.AddVertice(Vertice{ID: 4, Lat: -20.30, Lng: -40.32})
	g.AddArestaDirecionada(1, 2, 2, 3)
	g.AddArestaDirecionada(2, 3, 2, 3)
	g.AddArestaBidirecional(1, 4, 10, 12)
	g.AddArestaBidirecional(4, 3, 10, 12)
	g.SetMetadata("teste", time.Unix(1, 0).UTC())
	return g
}

func TestDijkstraEAEstrelaEncontramMesmoMenorCaminho(t *testing.T) {
	g := grafoDeTeste()
	dijkstra := g.Dijkstra(1, 3)
	aEstrela := g.AEstrela(1, 3)

	for nome, resultado := range map[string]ResultadoBusca{
		"dijkstra": dijkstra,
		"astar":    aEstrela,
	} {
		if len(resultado.Caminho) != 3 {
			t.Fatalf("%s: caminho tem %d vértices, esperado 3", nome, len(resultado.Caminho))
		}
		if math.Abs(resultado.DistanciaKm-4) > 1e-9 {
			t.Fatalf("%s: distância = %f, esperado 4", nome, resultado.DistanciaKm)
		}
		if resultado.DuracaoMin != 6 {
			t.Fatalf("%s: duração = %f, esperado 6", nome, resultado.DuracaoMin)
		}
	}
}

func TestArestaDirecionadaNaoCriaCaminhoInverso(t *testing.T) {
	g := NovoGrafo()
	g.AddVertice(Vertice{ID: 1})
	g.AddVertice(Vertice{ID: 2})
	g.AddArestaDirecionada(1, 2, 1, 1)
	if resultado := g.Dijkstra(2, 1); len(resultado.Caminho) != 0 {
		t.Fatalf("era esperado não existir caminho inverso; obtido %+v", resultado.Caminho)
	}
}

func TestSalvarECarregarArquivoPreservaGrafo(t *testing.T) {
	g := grafoDeTeste()
	caminho := filepath.Join(t.TempDir(), "grafo.gz")
	if err := g.SalvarArquivo(caminho); err != nil {
		t.Fatal(err)
	}
	recarregado, err := CarregarArquivo(caminho)
	if err != nil {
		t.Fatal(err)
	}
	if recarregado.QuantidadeVertices() != g.QuantidadeVertices() {
		t.Fatalf("vértices: obtido %d, esperado %d", recarregado.QuantidadeVertices(), g.QuantidadeVertices())
	}
	if recarregado.QuantidadeArestas() != g.QuantidadeArestas() {
		t.Fatalf("arestas: obtido %d, esperado %d", recarregado.QuantidadeArestas(), g.QuantidadeArestas())
	}
	resultado := recarregado.AEstrela(1, 3)
	if resultado.DistanciaKm != 4 {
		t.Fatalf("distância recarregada = %f, esperado 4", resultado.DistanciaKm)
	}
}
