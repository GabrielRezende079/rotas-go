// Package osmimport transforma vias do OpenStreetMap em um grafo pesquisável.
package osmimport

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"

	"rotas-go/internal/grafo"
	geo "rotas-go/internal/math"
)

// Relatorio resume o processo de conversão do PBF.
type Relatorio struct {
	ViasAceitas        int
	NosReferenciados   int
	NosCarregados      int
	Vertices           int
	Arestas            int64
	SegmentosIgnorados int
	Duracao            time.Duration
}

type viaOSM struct {
	nos           []int64
	direcao       int // 0: ambos, 1: ordem OSM, -1: ordem inversa
	velocidadeKmH float64
}

var classesPermitidas = map[string]float64{
	"motorway":       100,
	"motorway_link":  60,
	"trunk":          80,
	"trunk_link":     50,
	"primary":        70,
	"primary_link":   45,
	"secondary":      60,
	"secondary_link": 40,
	"tertiary":       50,
	"tertiary_link":  35,
	"unclassified":   40,
	"residential":    30,
	"living_street":  20,
	"service":        20,
	"road":           30,
}

// Construir lê um arquivo .osm.pbf e gera o grafo rodoviário para automóveis.
func Construir(ctx context.Context, caminhoPBF string) (*grafo.Grafo, Relatorio, error) {
	inicio := time.Now()
	vias, necessarios, err := lerVias(ctx, caminhoPBF)
	if err != nil {
		return nil, Relatorio{}, err
	}
	coordenadas, err := lerNos(ctx, caminhoPBF, necessarios)
	if err != nil {
		return nil, Relatorio{}, err
	}

	g := grafo.NovoGrafo()
	for id, coordenada := range coordenadas {
		g.AddVertice(grafo.Vertice{ID: id, Lat: coordenada[0], Lng: coordenada[1]})
	}

	ignorados := 0
	for _, via := range vias {
		for i := 0; i+1 < len(via.nos); i++ {
			origemID := via.nos[i]
			destinoID := via.nos[i+1]
			origem, okOrigem := coordenadas[origemID]
			destino, okDestino := coordenadas[destinoID]
			if !okOrigem || !okDestino || origemID == destinoID {
				ignorados++
				continue
			}
			distancia := geo.HaversineKm(origem[0], origem[1], destino[0], destino[1])
			if distancia <= 0 {
				ignorados++
				continue
			}
			duracao := distancia / via.velocidadeKmH * 60
			switch via.direcao {
			case 1:
				g.AddArestaDirecionada(origemID, destinoID, distancia, duracao)
			case -1:
				g.AddArestaDirecionada(destinoID, origemID, distancia, duracao)
			default:
				g.AddArestaBidirecional(origemID, destinoID, distancia, duracao)
			}
		}
	}

	g.SetMetadata(
		fmt.Sprintf("OpenStreetMap: %s", filepath.Base(caminhoPBF)),
		time.Now().UTC(),
	)
	relatorio := Relatorio{
		ViasAceitas:        len(vias),
		NosReferenciados:   len(necessarios),
		NosCarregados:      len(coordenadas),
		Vertices:           g.QuantidadeVertices(),
		Arestas:            g.QuantidadeArestas(),
		SegmentosIgnorados: ignorados,
		Duracao:            time.Since(inicio),
	}
	if relatorio.Vertices == 0 || relatorio.Arestas == 0 {
		return nil, relatorio, fmt.Errorf("nenhuma via dirigível foi encontrada em %s", caminhoPBF)
	}
	return g, relatorio, nil
}

func lerVias(ctx context.Context, caminho string) ([]viaOSM, map[int64]struct{}, error) {
	arquivo, err := os.Open(caminho)
	if err != nil {
		return nil, nil, err
	}
	defer arquivo.Close()

	scanner := osmpbf.New(ctx, arquivo, runtime.GOMAXPROCS(0))
	scanner.SkipNodes = true
	scanner.SkipRelations = true
	defer scanner.Close()

	vias := make([]viaOSM, 0, 100_000)
	necessarios := make(map[int64]struct{}, 500_000)
	for scanner.Scan() {
		way, ok := scanner.Object().(*osm.Way)
		if !ok || !viaPermitida(way) {
			continue
		}
		nos := make([]int64, len(way.Nodes))
		for i, node := range way.Nodes {
			id := int64(node.ID)
			nos[i] = id
			necessarios[id] = struct{}{}
		}
		classe := way.Tags.Find("highway")
		vias = append(vias, viaOSM{
			nos:           nos,
			direcao:       direcaoVia(way, classe),
			velocidadeKmH: velocidadeVia(way, classe),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("lendo vias do PBF: %w", err)
	}
	return vias, necessarios, nil
}

func lerNos(ctx context.Context, caminho string, necessarios map[int64]struct{}) (map[int64][2]float64, error) {
	arquivo, err := os.Open(caminho)
	if err != nil {
		return nil, err
	}
	defer arquivo.Close()

	scanner := osmpbf.New(ctx, arquivo, runtime.GOMAXPROCS(0))
	scanner.SkipWays = true
	scanner.SkipRelations = true
	scanner.FilterNode = func(node *osm.Node) bool {
		_, ok := necessarios[int64(node.ID)]
		return ok
	}
	defer scanner.Close()

	coordenadas := make(map[int64][2]float64, len(necessarios))
	for scanner.Scan() {
		node, ok := scanner.Object().(*osm.Node)
		if !ok {
			continue
		}
		coordenadas[int64(node.ID)] = [2]float64{node.Lat, node.Lon}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("lendo nós do PBF: %w", err)
	}
	return coordenadas, nil
}

func viaPermitida(way *osm.Way) bool {
	if len(way.Nodes) < 2 {
		return false
	}
	classe := way.Tags.Find("highway")
	if _, ok := classesPermitidas[classe]; !ok {
		return false
	}
	for _, chave := range []string{"access", "vehicle", "motor_vehicle", "motorcar"} {
		valor := strings.ToLower(way.Tags.Find(chave))
		if valor == "no" || valor == "private" {
			return false
		}
	}
	return true
}

func direcaoVia(way *osm.Way, classe string) int {
	valor := strings.ToLower(strings.TrimSpace(way.Tags.Find("oneway")))
	switch valor {
	case "yes", "true", "1":
		return 1
	case "-1", "reverse":
		return -1
	case "no", "false", "0":
		return 0
	}
	if way.Tags.Find("junction") == "roundabout" || classe == "motorway" || classe == "motorway_link" {
		return 1
	}
	return 0
}

func velocidadeVia(way *osm.Way, classe string) float64 {
	padrao := classesPermitidas[classe]
	valor := strings.ToLower(strings.TrimSpace(way.Tags.Find("maxspeed")))
	if valor == "" {
		return padrao
	}
	if indice := strings.Index(valor, ";"); indice >= 0 {
		valor = valor[:indice]
	}
	numero := prefixoNumerico(valor)
	if numero == "" {
		return padrao
	}
	velocidade, err := strconv.ParseFloat(strings.ReplaceAll(numero, ",", "."), 64)
	if err != nil || velocidade <= 0 {
		return padrao
	}
	if strings.Contains(valor, "mph") {
		velocidade *= 1.609344
	}
	return velocidade
}

func prefixoNumerico(valor string) string {
	valor = strings.TrimSpace(valor)
	fim := 0
	for fim < len(valor) {
		c := valor[fim]
		if (c < '0' || c > '9') && c != '.' && c != ',' {
			break
		}
		fim++
	}
	return valor[:fim]
}
