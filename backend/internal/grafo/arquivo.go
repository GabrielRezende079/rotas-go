package grafo

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const versaoArquivoGrafo = 1

type cabecalhoArquivo struct {
	Versao   int
	Fonte    string
	GeradoEm time.Time
	Vertices int
	Arestas  int64
}

type arestaArquivo struct {
	Origem      int64
	Destino     int64
	DistanciaKm float64
	DuracaoMin  float64
}

// SalvarArquivo serializa o grafo em GOB comprimido, sem duplicar toda a malha
// em memória. A troca do arquivo é atômica para evitar caches incompletos.
func (g *Grafo) SalvarArquivo(caminho string) (errRetorno error) {
	if err := os.MkdirAll(filepath.Dir(caminho), 0o755); err != nil {
		return err
	}
	temporario := caminho + ".tmp"
	defer func() {
		if errRetorno != nil {
			_ = os.Remove(temporario)
		}
	}()

	arquivo, err := os.Create(temporario)
	if err != nil {
		return err
	}
	zip := gzip.NewWriter(arquivo)
	encoder := gob.NewEncoder(zip)

	meta := g.Metadata()
	cabecalho := cabecalhoArquivo{
		Versao:   versaoArquivoGrafo,
		Fonte:    meta.Fonte,
		GeradoEm: meta.GeradoEm,
		Vertices: meta.Vertices,
		Arestas:  meta.Arestas,
	}
	if err := encoder.Encode(cabecalho); err != nil {
		_ = zip.Close()
		_ = arquivo.Close()
		return err
	}
	for _, vertice := range g.vertices {
		if err := encoder.Encode(vertice); err != nil {
			_ = zip.Close()
			_ = arquivo.Close()
			return err
		}
	}
	for origem, arestas := range g.adjacencias {
		for _, aresta := range arestas {
			registro := arestaArquivo{
				Origem:      origem,
				Destino:     aresta.Destino,
				DistanciaKm: aresta.DistanciaKm,
				DuracaoMin:  aresta.DuracaoMin,
			}
			if err := encoder.Encode(registro); err != nil {
				_ = zip.Close()
				_ = arquivo.Close()
				return err
			}
		}
	}
	if err := zip.Close(); err != nil {
		_ = arquivo.Close()
		return err
	}
	if err := arquivo.Close(); err != nil {
		return err
	}
	return os.Rename(temporario, caminho)
}

// CarregarArquivo restaura um grafo previamente criado pelo importador.
func CarregarArquivo(caminho string) (*Grafo, error) {
	arquivo, err := os.Open(caminho)
	if err != nil {
		return nil, err
	}
	defer arquivo.Close()
	zip, err := gzip.NewReader(arquivo)
	if err != nil {
		return nil, err
	}
	defer zip.Close()
	decoder := gob.NewDecoder(zip)

	var cabecalho cabecalhoArquivo
	if err := decoder.Decode(&cabecalho); err != nil {
		return nil, err
	}
	if cabecalho.Versao != versaoArquivoGrafo {
		return nil, fmt.Errorf("versão de grafo incompatível: %d", cabecalho.Versao)
	}
	if cabecalho.Vertices < 1 || cabecalho.Arestas < 1 {
		return nil, fmt.Errorf("arquivo de grafo vazio")
	}

	g := &Grafo{
		vertices:    make(map[int64]Vertice, cabecalho.Vertices),
		adjacencias: make(map[int64][]Aresta),
		metadata: Metadata{
			Fonte:    cabecalho.Fonte,
			GeradoEm: cabecalho.GeradoEm,
		},
	}
	for i := 0; i < cabecalho.Vertices; i++ {
		var vertice Vertice
		if err := decoder.Decode(&vertice); err != nil {
			return nil, fmt.Errorf("lendo vértice %d: %w", i, err)
		}
		g.AddVertice(vertice)
	}
	for i := int64(0); i < cabecalho.Arestas; i++ {
		var registro arestaArquivo
		if err := decoder.Decode(&registro); err != nil {
			return nil, fmt.Errorf("lendo aresta %d: %w", i, err)
		}
		g.AddArestaDirecionada(
			registro.Origem,
			registro.Destino,
			registro.DistanciaKm,
			registro.DuracaoMin,
		)
	}
	return g, nil
}
