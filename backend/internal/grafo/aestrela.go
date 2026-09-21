package grafo

import (
	"container/heap"

	geo "rotas-go/internal/math"
)

// AEstrela encontra o caminho mínimo com A*; a heurística é a distância em
// linha reta (Haversine) até o destino — consistente, então o caminho
// encontrado é o mesmo do Dijkstra.
func (g *Grafo) AEstrela(origemID, destinoID string) ResultadoBusca {
	origem, ok := g.vertices[origemID]
	if !ok {
		return ResultadoBusca{}
	}
	destino, ok := g.vertices[destinoID]
	if !ok {
		return ResultadoBusca{}
	}

	custoG := make(map[string]float64)
	custoG[origemID] = 0
	anterior := make(map[string]string)
	fechados := make(map[string]bool)

	fila := &filaPrioridade{}
	heap.Init(fila)
	heap.Push(fila, &itemPQ{
		id:         origemID,
		prioridade: geo.HaversineKm(origem.Lat, origem.Lng, destino.Lat, destino.Lng),
	})

	visitados := 0

	for fila.Len() > 0 {
		atual := heap.Pop(fila).(*itemPQ)
		visitados++
		if atual.id == destinoID {
			break
		}
		if fechados[atual.id] {
			continue
		}
		fechados[atual.id] = true

		for _, aresta := range g.adjacencias[atual.id] {
			novoG := custoG[atual.id] + aresta.PesoKm
			if atualG, ok := custoG[aresta.Destino]; !ok || novoG < atualG {
				custoG[aresta.Destino] = novoG
				anterior[aresta.Destino] = atual.id
				vizinho := g.vertices[aresta.Destino]
				h := geo.HaversineKm(vizinho.Lat, vizinho.Lng, destino.Lat, destino.Lng)
				heap.Push(fila, &itemPQ{id: aresta.Destino, prioridade: novoG + h})
			}
		}
	}

	if _, ok := custoG[destinoID]; !ok {
		return ResultadoBusca{}
	}

	return ResultadoBusca{
		Caminho:        reconstruirCaminho(g, anterior, origemID, destinoID),
		DistanciaKm:    custoG[destinoID],
		NodosVisitados: visitados,
	}
}
