package grafo

import (
	"container/heap"

	geo "rotas-go/internal/math"
)

// vMaxKmh é a maior velocidade assumida nas vias; usada para escalar a
// heurística quando o custo é a duração, mantendo a admissibilidade do A*.
const vMaxKmh = 130.0

// AEstrela encontra a menor distância usando Haversine como limite inferior.
// A heurística permanece em linha reta, mas o caminho retornado segue as arestas
// reais da malha viária importada do OpenStreetMap.
func (g *Grafo) AEstrela(origemID, destinoID int64) ResultadoBusca {
	return g.buscaAEstrela(origemID, destinoID, false)
}

// AEstrelaPorDuracao encontra o caminho mais rápido usando uma heurística
// escalada para minutos, ainda admissível em relação ao custo por duração.
func (g *Grafo) AEstrelaPorDuracao(origemID, destinoID int64) ResultadoBusca {
	return g.buscaAEstrela(origemID, destinoID, true)
}

// AEstrelaComRestricoes é a variante do A* que respeita nós e arestas proibidos,
// usada como spur path no cálculo de alternativas (algoritmo de Yen).
func (g *Grafo) AEstrelaComRestricoes(
	origemID, destinoID int64,
	porDuracao bool,
	ignorarNos map[int64]bool,
	ignorarArestas map[[2]int64]bool,
) ResultadoBusca {
	return g.buscaAEstrelaParametrizada(origemID, destinoID, porDuracao, ignorarNos, ignorarArestas)
}

func (g *Grafo) buscaAEstrela(origemID, destinoID int64, porDuracao bool) ResultadoBusca {
	return g.buscaAEstrelaParametrizada(origemID, destinoID, porDuracao, nil, nil)
}

func (g *Grafo) buscaAEstrelaParametrizada(
	origemID, destinoID int64,
	porDuracao bool,
	ignorarNos map[int64]bool,
	ignorarArestas map[[2]int64]bool,
) ResultadoBusca {
	origem, ok := g.vertices[origemID]
	if !ok {
		return ResultadoBusca{}
	}
	destino, ok := g.vertices[destinoID]
	if !ok {
		return ResultadoBusca{}
	}

	heuristica := func(lat, lng float64) float64 {
		km := geo.HaversineKm(lat, lng, destino.Lat, destino.Lng)
		if porDuracao {
			return km * 60.0 / vMaxKmh
		}
		return km
	}
	custoDe := func(a Aresta) float64 {
		if porDuracao {
			return a.DuracaoMin
		}
		return a.DistanciaKm
	}

	custoG := map[int64]float64{origemID: 0}
	anterior := make(map[int64]passoAnterior)
	fila := &filaPrioridade{}
	heap.Init(fila)
	heap.Push(fila, &itemPQ{
		id:         origemID,
		prioridade: heuristica(origem.Lat, origem.Lng),
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
			if ignorarNos != nil && ignorarNos[aresta.Destino] {
				continue
			}
			if ignorarArestas != nil && ignorarArestas[[2]int64{atual.id, aresta.Destino}] {
				continue
			}
			novoG := melhor + custoDe(aresta)
			antigo, conhecido := custoG[aresta.Destino]
			if conhecido && novoG >= antigo {
				continue
			}
			custoG[aresta.Destino] = novoG
			anterior[aresta.Destino] = passoAnterior{origem: atual.id, aresta: aresta}
			vizinho := g.vertices[aresta.Destino]
			f := novoG + heuristica(vizinho.Lat, vizinho.Lng)
			heap.Push(fila, &itemPQ{id: aresta.Destino, prioridade: f, custo: novoG})
		}
	}

	if _, ok := custoG[destinoID]; !ok {
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
