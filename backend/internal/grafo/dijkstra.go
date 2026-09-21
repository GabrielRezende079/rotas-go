package grafo

import (
	"container/heap"
	"math"
)

// itemPQ representa um vértice na fila de prioridade.
type itemPQ struct {
	id         string
	prioridade float64
	indice     int
}

// filaPrioridade implementa heap.Interface para usar com container/heap.
type filaPrioridade []*itemPQ

func (f filaPrioridade) Len() int { return len(f) }

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

// reconstruirCaminho monta o caminho origem → destino a partir dos predecessores.
func reconstruirCaminho(g *Grafo, anterior map[string]string, origemID, destinoID string) []*Vertice {
	caminho := []*Vertice{g.vertices[destinoID]}
	idAtual := destinoID
	for idAtual != origemID {
		caminho = append(caminho, g.vertices[anterior[idAtual]])
		idAtual = anterior[idAtual]
	}
	for i, j := 0, len(caminho)-1; i < j; i, j = i+1, j-1 {
		caminho[i], caminho[j] = caminho[j], caminho[i]
	}
	return caminho
}

// Dijkstra encontra o caminho de menor custo entre dois vértices.
func (g *Grafo) Dijkstra(origemID, destinoID string) ResultadoBusca {
	if _, ok := g.vertices[origemID]; !ok {
		return ResultadoBusca{}
	}
	if _, ok := g.vertices[destinoID]; !ok {
		return ResultadoBusca{}
	}

	const infinito = math.MaxFloat64

	distancias := make(map[string]float64, len(g.vertices))
	for id := range g.vertices {
		distancias[id] = infinito
	}
	distancias[origemID] = 0

	anterior := make(map[string]string)

	fila := &filaPrioridade{}
	heap.Init(fila)
	heap.Push(fila, &itemPQ{id: origemID, prioridade: 0})

	visitados := 0

	for fila.Len() > 0 {
		atual := heap.Pop(fila).(*itemPQ)
		visitados++
		if atual.id == destinoID {
			break
		}
		if atual.prioridade > distancias[atual.id] {
			continue // entrada obsoleta na fila
		}
		for _, aresta := range g.adjacencias[atual.id] {
			if novo := distancias[atual.id] + aresta.PesoKm; novo < distancias[aresta.Destino] {
				distancias[aresta.Destino] = novo
				anterior[aresta.Destino] = atual.id
				heap.Push(fila, &itemPQ{id: aresta.Destino, prioridade: novo})
			}
		}
	}

	if distancias[destinoID] == infinito {
		return ResultadoBusca{}
	}

	return ResultadoBusca{
		Caminho:        reconstruirCaminho(g, anterior, origemID, destinoID),
		DistanciaKm:    distancias[destinoID],
		NodosVisitados: visitados,
	}
}
