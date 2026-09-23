package grafo

import (
	"container/heap"
)

const (
	// dissimilaridadeMinima é a fração mínima de vértices diferentes que uma
	// alternativa precisa exibir em relação a qualquer rota já aceita.
	// Em malhas viárias reais as alternativas compartilham o tronco principal e
	// a convergência nos extremos; cerca de 5% de vértices distintos já
	// representa um desvio de trecho perceptível no mapa.
	dissimilaridadeMinima = 0.05
	// maxTermosPorCaminho limita quantos pontos de desvio o Yen expande por
	// caminho base. Desvios tardios geram as variantes perceptíveis.
	maxTermosPorCaminho = 40
	// limiteCandidatosYen limita o número total de buscas spur por trecho,
	// mantendo o custo computacional previsível no grafo viário real.
	limiteCandidatosYen = 14
)

// caminhoEncontrado guarda um caminho e seus custos para o algoritmo de Yen.
type caminhoEncontrado struct {
	nos          []int64
	vertices     []Vertice
	arestas      []Aresta
	custo        float64
	distKm       float64
	duracaoMin   float64
	visitados    int
	prefixoCusto []float64
}

func caminhoDeBusca(r ResultadoBusca) *caminhoEncontrado {
	nos := make([]int64, len(r.Caminho))
	for i, v := range r.Caminho {
		nos[i] = v.ID
	}
	return &caminhoEncontrado{
		nos:          nos,
		vertices:     r.Caminho,
		arestas:      r.Arestas,
		custo:        r.DuracaoMin,
		distKm:       r.DistanciaKm,
		duracaoMin:   r.DuracaoMin,
		visitados:    r.NodosVisitados,
		prefixoCusto: prefixosDeCusto(r.Arestas),
	}
}

func prefixosDeCusto(arestas []Aresta) []float64 {
	pref := make([]float64, len(arestas)+1)
	for i, a := range arestas {
		pref[i+1] = pref[i] + a.DuracaoMin
	}
	return pref
}

type candidatoYen struct {
	caminho *caminhoEncontrado
}

type filaCandidatos []*candidatoYen

func (f filaCandidatos) Len() int           { return len(f) }
func (f filaCandidatos) Less(i, j int) bool { return f[i].caminho.custo < f[j].caminho.custo }
func (f filaCandidatos) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f *filaCandidatos) Push(x any)        { *f = append(*f, x.(*candidatoYen)) }
func (f *filaCandidatos) Pop() any {
	antigo := *f
	n := len(antigo)
	c := antigo[n-1]
	antigo[n-1] = nil
	*f = antigo[:n-1]
	return c
}

// KMenoresCaminhos devolve até maxAlternativas rotas alternativas entre dois
// nós, ordenadas da mais rápida para a mais lenta pela duração estimada.
//
// Usa o algoritmo de Yen com filtro de dissimilaridade: o caminho mais rápido
// é sempre devolvido; os demais só são aceitos se forem suficientemente
// diferentes dos já aceitos (critério de Sørensen-Dice sobre os vértices).
func (g *Grafo) KMenoresCaminhos(origemID, destinoID int64, maxAlternativas int) []ResultadoBusca {
	if maxAlternativas < 1 {
		maxAlternativas = 3
	}
	if maxAlternativas > 3 {
		maxAlternativas = 3
	}

	primeira := g.dijkstraParametrizado(origemID, destinoID, true, nil, nil)
	if len(primeira.Caminho) == 0 {
		return nil
	}
	if maxAlternativas == 1 || len(primeira.Caminho) < 3 {
		return []ResultadoBusca{primeira}
	}

	base := caminhoDeBusca(primeira)
	aceitas := [][]int64{base.nos}
	resultados := []ResultadoBusca{primeira}
	fila := &filaCandidatos{}
	heap.Init(fila)
	gerados := 0

	g.gerarCandidatosYen(base, destinoID, aceitas, fila, &gerados)

	for len(resultados) < maxAlternativas && fila.Len() > 0 {
		cand := heap.Pop(fila).(*candidatoYen)
		if !dissimilarO(cand.caminho.nos, aceitas) {
			continue
		}
		aceitas = append(aceitas, cand.caminho.nos)
		resultados = append(resultados, resultadoDeCaminho(cand.caminho))
		g.gerarCandidatosYen(cand.caminho, destinoID, aceitas, fila, &gerados)
	}
	return resultados
}

// gerarCandidatosYen calcula os desvios (spur) do caminho base e enfileira os
// candidatos compostos, respeitando os limites de termos e de buscas spur.
// temSaidaViavel indica se o nó possui alguma aresta de saída que atravessa as
// restrições de nós e arestas proibidos (usada para podar spurs impossíveis).
func (g *Grafo) temSaidaViavel(nos int64, ignorarNos map[int64]bool, ignorarArestas map[[2]int64]bool) bool {
	for _, aresta := range g.adjacencias[nos] {
		if ignorarNos != nil && ignorarNos[aresta.Destino] {
			continue
		}
		if ignorarArestas != nil && ignorarArestas[[2]int64{nos, aresta.Destino}] {
			continue
		}
		return true
	}
	return false
}

func (g *Grafo) gerarCandidatosYen(
	base *caminhoEncontrado,
	destinoID int64,
	aceitas [][]int64,
	fila *filaCandidatos,
	gerados *int,
) {
	maxI := len(base.nos) - 1
	if limite := maxTermosPorCaminho; limite < maxI {
		maxI = limite
	}
	for i := 0; i < maxI; i++ {
		if *gerados >= limiteCandidatosYen {
			return
		}
		ignorarNos := make(map[int64]bool, i)
		for j := 0; j < i; j++ {
			ignorarNos[base.nos[j]] = true
		}
		ignorarArestas := make(map[[2]int64]bool)
		for _, aceita := range aceitas {
			if i+1 < len(aceita) {
				ignorarArestas[[2]int64{aceita[i], aceita[i+1]}] = true
			}
		}

		// Poda barata: se o nó de desvio não tem nenhuma aresta de saída que
		// sobreviva às restrições, o spur seria uma busca inteira inútil.
		if !g.temSaidaViavel(base.nos[i], ignorarNos, ignorarArestas) {
			continue
		}

		spur := g.AEstrelaComRestricoes(base.nos[i], destinoID, true, ignorarNos, ignorarArestas)
		if len(spur.Caminho) == 0 {
			continue
		}
		spurPath := caminhoDeBusca(spur)

		nos := make([]int64, 0, i+1+len(spurPath.nos)-1)
		nos = append(nos, base.nos[:i+1]...)
		nos = append(nos, spurPath.nos[1:]...)
		vertices := make([]Vertice, 0, len(nos))
		vertices = append(vertices, base.vertices[:i+1]...)
		vertices = append(vertices, spurPath.vertices[1:]...)
		arestas := make([]Aresta, 0, i+len(spurPath.arestas))
		arestas = append(arestas, base.arestas[:i]...)
		arestas = append(arestas, spurPath.arestas...)

		var dist, dur float64
		for _, a := range arestas {
			dist += a.DistanciaKm
			dur += a.DuracaoMin
		}
		composto := &caminhoEncontrado{
			nos:          nos,
			vertices:     vertices,
			arestas:      arestas,
			custo:        dur,
			distKm:       dist,
			duracaoMin:   dur,
			visitados:    base.visitados + spurPath.visitados,
			prefixoCusto: prefixosDeCusto(arestas),
		}
		*gerados++
		heap.Push(fila, &candidatoYen{caminho: composto})
	}
}

// dissimilarO verifica se o candidato difere suficientemente de todas as rotas
// já aceitas.
func dissimilarO(nos []int64, aceitas [][]int64) bool {
	for _, aceita := range aceitas {
		if similaridadeNos(nos, aceita) >= 1-dissimilaridadeMinima {
			return false
		}
	}
	return true
}

// similaridadeNos calcula o coeficiente de Sørensen-Dice entre dois conjuntos
// de vértices: 2*|A∩B| / (|A|+|B|).
func similaridadeNos(a, b []int64) float64 {
	conj := make(map[int64]struct{}, len(a))
	for _, n := range a {
		conj[n] = struct{}{}
	}
	intersecao := 0
	for _, n := range b {
		if _, ok := conj[n]; ok {
			intersecao++
		}
	}
	return 2 * float64(intersecao) / float64(len(a)+len(b))
}

func resultadoDeCaminho(p *caminhoEncontrado) ResultadoBusca {
	return ResultadoBusca{
		Caminho:        p.vertices,
		Arestas:        p.arestas,
		DistanciaKm:    p.distKm,
		DuracaoMin:     p.duracaoMin,
		NodosVisitados: p.visitados,
	}
}
