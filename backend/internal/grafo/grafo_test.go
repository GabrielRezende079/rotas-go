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

func TestDijkstraPorDuracaoEscolheCaminhoRapido(t *testing.T) {
	g := NovoGrafo()
	g.AddVertice(Vertice{ID: 1})
	g.AddVertice(Vertice{ID: 2})
	g.AddVertice(Vertice{ID: 3})
	// Caminho curto em distância, mas lento.
	g.AddArestaDirecionada(1, 3, 1, 60)
	// Caminho longo em distância, mas rápido.
	g.AddArestaDirecionada(1, 2, 15, 5)
	g.AddArestaDirecionada(2, 3, 15, 5)

	porDistancia := g.Dijkstra(1, 3)
	if porDistancia.DistanciaKm != 1 {
		t.Fatalf("por distância: distância = %f, esperado 1", porDistancia.DistanciaKm)
	}
	if porDistancia.DuracaoMin != 60 {
		t.Fatalf("por distância: duração = %f, esperado 60", porDistancia.DuracaoMin)
	}

	porDuracao := g.DijkstraPorDuracao(1, 3)
	if porDuracao.DuracaoMin != 10 {
		t.Fatalf("por duração: duração = %f, esperado 10", porDuracao.DuracaoMin)
	}
	if porDuracao.DistanciaKm != 30 {
		t.Fatalf("por duração: distância = %f, esperado 30", porDuracao.DistanciaKm)
	}
}

func TestAEstrelaPorDuracaoUsaMesmoCusto(t *testing.T) {
	g := NovoGrafo()
	g.AddVertice(Vertice{ID: 1, Lat: -20.0, Lng: -40.0})
	g.AddVertice(Vertice{ID: 2, Lat: -20.0, Lng: -40.001})
	g.AddVertice(Vertice{ID: 3, Lat: -20.0001, Lng: -40.001})
	g.AddArestaDirecionada(1, 2, 1, 2)
	g.AddArestaDirecionada(2, 3, 1, 2)

	d := g.DijkstraPorDuracao(1, 3)
	a := g.AEstrelaPorDuracao(1, 3)
	if math.Abs(d.DuracaoMin-a.DuracaoMin) > 1e-9 {
		t.Fatalf("durações divergem: dijkstra=%f astar=%f", d.DuracaoMin, a.DuracaoMin)
	}
	if math.Abs(d.DistanciaKm-a.DistanciaKm) > 1e-9 {
		t.Fatalf("distâncias divergem: dijkstra=%f astar=%f", d.DistanciaKm, a.DistanciaKm)
	}
}

func TestKMenoresCaminhosRetornaAlternativasDissimilaridade(t *testing.T) {
	g := NovoGrafo()
	for id := 1; id <= 10; id++ {
		g.AddVertice(Vertice{ID: int64(id)})
	}
	// Corredor 1 (mais rápido, duração 2 por aresta).
	for i := 1; i <= 5; i++ {
		g.AddArestaBidirecional(int64(i), int64(i+1), 1, 2)
	}
	g.AddArestaBidirecional(6, 10, 1, 2)
	// Corredor 2 (mais lento, duração 4 por aresta).
	for _, par := range [][2]int64{{1, 7}, {7, 8}, {8, 9}, {9, 10}} {
		g.AddArestaBidirecional(par[0], par[1], 10, 4)
	}

	resultados := g.KMenoresCaminhos(1, 10, 3)
	if len(resultados) < 2 {
		t.Fatalf("esperava ao menos 2 alternativas, obtidas %d", len(resultados))
	}
	primeira := resultados[0]
	if len(primeira.Caminho) != 7 {
		t.Fatalf("melhor rota deveria ser o corredor 1 (7 vértices), obtidos %d", len(primeira.Caminho))
	}
	for i := 1; i < len(resultados); i++ {
		if resultados[i].DuracaoMin < resultados[i-1].DuracaoMin {
			t.Fatalf("alternativas fora de ordem de duração: %f depois %f", resultados[i].DuracaoMin, resultados[i-1].DuracaoMin)
		}
		for _, anterior := range resultados[:i] {
			if s := similaridadeNos(resultados[i].nosDeCaminho(), anterior.nosDeCaminho()); s >= 1-dissimilaridadeMinima {
				t.Fatalf("alternativa %d é similar demais (Sørensen-Dice=%.2f)", i, s)
			}
		}
	}
}

func (r ResultadoBusca) nosDeCaminho() []int64 {
	nos := make([]int64, len(r.Caminho))
	for i, v := range r.Caminho {
		nos[i] = v.ID
	}
	return nos
}

func TestKMenoresCaminhosOrdenaPorCustoEContaMetricas(t *testing.T) {
	g := NovoGrafo()
	for id := 1; id <= 7; id++ {
		g.AddVertice(Vertice{ID: int64(id)})
	}
	g.AddArestaBidirecional(1, 2, 1, 1)
	g.AddArestaBidirecional(2, 3, 1, 1)
	g.AddArestaBidirecional(3, 4, 1, 1)
	g.AddArestaBidirecional(4, 5, 1, 1)
	g.AddArestaBidirecional(5, 6, 1, 1)
	g.AddArestaDirecionada(6, 7, 1, 1)
	g.AddArestaBidirecional(1, 6, 3, 9)
	g.AddArestaDirecionada(4, 7, 1, 1)

	resultados := g.KMenoresCaminhos(1, 7, 3)
	if len(resultados) == 0 {
		t.Fatalf("nenhum caminho encontrado")
	}
	primeiro := resultados[0]
	if primeiro.DistanciaKm <= 0 || primeiro.DuracaoMin <= 0 {
		t.Fatalf("métricas não preenchidas: dist=%.2f dur=%.2f", primeiro.DistanciaKm, primeiro.DuracaoMin)
	}
	for i := 1; i < len(resultados); i++ {
		if resultados[i].DuracaoMin < resultados[i-1].DuracaoMin {
			t.Fatalf("alternativas fora de ordem de duração: %f depois %f", resultados[i].DuracaoMin, resultados[i-1].DuracaoMin)
		}
	}
}

func TestDijkstraParametrizadoRespeitaArestasIgnoradas(t *testing.T) {
	g := grafoDeTeste()
	// 1-2-3 é o caminho rápido; bloquear a aresta (2->3) força o caminho 1-4-3.
	resultado := g.dijkstraParametrizado(1, 3, true, nil, map[[2]int64]bool{{2, 3}: true})
	if len(resultado.Caminho) == 0 {
		t.Fatalf("nenhum caminho encontrado")
	}
	if resultado.DistanciaKm != 20 {
		t.Fatalf("distância = %f, esperado 20", resultado.DistanciaKm)
	}
}
