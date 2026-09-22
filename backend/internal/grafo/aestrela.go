package grafo

import (
	"container/heap"

	geo "rotas-go/internal/math"
)

// AEstrela encontra a menor distância usando Haversine como limite inferior.
// A heurística permanece em linha reta, mas o caminho retornado segue as arestas
// reais da malha viária importada do OpenStreetMap.
func (g *Grafo) AEstrela(origemID, destinoID int64) ResultadoBusca {
	origem, ok := g.vertices[origemID]
	if !ok {
		return ResultadoBusca{}
	}
	destino, ok := g.vertices[destinoID]
	if !ok {
		return ResultadoBusca{}
	}

	custoG := map[int64]float64{origemID: 0}
	anterior := make(map[int64]passoAnterior)
	fila := &filaPrioridade{}
	heap.Init(fila)
	heap.Push(fila, &itemPQ{
		id:         origemID,
		prioridade: geo.HaversineKm(origem.Lat, origem.Lng, destino.Lat, destino.Lng),
		custo:      0,
	})
	visitados := 0

	for fila.Len() > 0 {
		atual := heap.Pop(fila).(*itemPQ)
		melhor, existe := custoG[atual.id]
		if !existe || atual.custo > melhor {
			continue
		}
		visitados++
		if atual.id == destinoID {
			break
		}
		for _, aresta := range g.adjacencias[atual.id] {
			novoG := melhor + aresta.DistanciaKm
			antigo, conhecido := custoG[aresta.Destino]
			if conhecido && novoG >= antigo {
				continue
			}
			custoG[aresta.Destino] = novoG
			anterior[aresta.Destino] = passoAnterior{origem: atual.id, aresta: aresta}
			vizinho := g.vertices[aresta.Destino]
			h := geo.HaversineKm(vizinho.Lat, vizinho.Lng, destino.Lat, destino.Lng)
			heap.Push(fila, &itemPQ{id: aresta.Destino, prioridade: novoG + h, custo: novoG})
		}
	}

	distancia, ok := custoG[destinoID]
	if !ok {
		return ResultadoBusca{}
	}
	resultado := reconstruirResultado(g, anterior, origemID, destinoID)
	resultado.DistanciaKm = distancia
	resultado.NodosVisitados = visitados
	for _, aresta := range resultado.Arestas {
		resultado.DuracaoMin += aresta.DuracaoMin
	}
	return resultado
}
