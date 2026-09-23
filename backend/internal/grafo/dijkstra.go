package grafo

import (
	"container/heap"
	"math"
)

type itemPQ struct {
	id         int64
	prioridade float64
	custo      float64
	indice     int
}

type filaPrioridade []*itemPQ

func (f filaPrioridade) Len() int           { return len(f) }
func (f filaPrioridade) Less(i, j int) bool { return f[i].prioridade < f[j].prioridade }
func (f filaPrioridade) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
	f[i].indice = i
	f[j].indice = j
}
func (f *filaPrioridade) Push(x any) {
	item := x.(*itemPQ)
	item.indice = len(*f)
	*f = append(*f, item)
}
func (f *filaPrioridade) Pop() any {
	antigo := *f
	n := len(antigo)
	item := antigo[n-1]
	antigo[n-1] = nil
	*f = antigo[:n-1]
	return item
}

type passoAnterior struct {
	origem int64
	aresta Aresta
}

func reconstruirResultado(g *Grafo, anterior map[int64]passoAnterior, origemID, destinoID int64) ResultadoBusca {
	destino, ok := g.vertices[destinoID]
	if !ok {
		return ResultadoBusca{}
	}
	vertices := []Vertice{destino}
	arestas := make([]Aresta, 0)
	atual := destinoID
	for atual != origemID {
		passo, existe := anterior[atual]
		if !existe {
			return ResultadoBusca{}
		}
		arestas = append(arestas, passo.aresta)
		atual = passo.origem
		vertices = append(vertices, g.vertices[atual])
	}
	for i, j := 0, len(vertices)-1; i < j; i, j = i+1, j-1 {
		vertices[i], vertices[j] = vertices[j], vertices[i]
	}
	for i, j := 0, len(arestas)-1; i < j; i, j = i+1, j-1 {
		arestas[i], arestas[j] = arestas[j], arestas[i]
	}
	return ResultadoBusca{Caminho: vertices, Arestas: arestas}
}

// Dijkstra encontra a menor distância viária entre dois nós.
func (g *Grafo) Dijkstra(origemID, destinoID int64) ResultadoBusca {
	return g.dijkstraParametrizado(origemID, destinoID, false, nil, nil)
}

// DijkstraPorDuracao encontra o caminho mais rápido entre dois nós.
func (g *Grafo) DijkstraPorDuracao(origemID, destinoID int64) ResultadoBusca {
	return g.dijkstraParametrizado(origemID, destinoID, true, nil, nil)
}

// dijkstraParametrizado executa Dijkstra com custo opcional por duração e com
// suporte a conjuntos de nós e arestas proibidos (usado pelo k-ésimos caminhos).
func (g *Grafo) dijkstraParametrizado(
	origemID, destinoID int64,
	porDuracao bool,
	ignorarNos map[int64]bool,
	ignorarArestas map[[2]int64]bool,
) ResultadoBusca {
	if _, ok := g.vertices[origemID]; !ok {
		return ResultadoBusca{}
	}
	if _, ok := g.vertices[destinoID]; !ok {
		return ResultadoBusca{}
	}

	custoDe := func(a Aresta) float64 {
		if porDuracao {
			return a.DuracaoMin
		}
		return a.DistanciaKm
	}

	custos := map[int64]float64{origemID: 0}
	anterior := make(map[int64]passoAnterior)
	fila := &filaPrioridade{}
	heap.Init(fila)
	heap.Push(fila, &itemPQ{id: origemID, prioridade: 0, custo: 0})
	visitados := 0

	for fila.Len() > 0 {
		atual := heap.Pop(fila).(*itemPQ)
		melhor, existe := custos[atual.id]
		if !existe || atual.custo > melhor {
			continue
		}
		visitados++
		if atual.id == destinoID {
			break
		}
		for _, aresta := range g.adjacencias[atual.id] {
			if ignorarNos != nil && ignorarNos[aresta.Destino] {
				continue
			}
			if ignorarArestas != nil && ignorarArestas[[2]int64{atual.id, aresta.Destino}] {
				continue
			}
			novo := melhor + custoDe(aresta)
			anteriorCusto, conhecido := custos[aresta.Destino]
			if !conhecido {
				anteriorCusto = math.Inf(1)
			}
			if novo < anteriorCusto {
				custos[aresta.Destino] = novo
				anterior[aresta.Destino] = passoAnterior{origem: atual.id, aresta: aresta}
				heap.Push(fila, &itemPQ{id: aresta.Destino, prioridade: novo, custo: novo})
			}
		}
	}

	if _, ok := custos[destinoID]; !ok {
		return ResultadoBusca{}
	}
	resultado := reconstruirResultado(g, anterior, origemID, destinoID)
	resultado.NodosVisitados = visitados
	for _, aresta := range resultado.Arestas {
		resultado.DistanciaKm += aresta.DistanciaKm
		resultado.DuracaoMin += aresta.DuracaoMin
	}
	return resultado
}
