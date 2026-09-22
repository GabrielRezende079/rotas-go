package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"rotas-go/internal/osmimport"
)

func main() {
	entrada := flag.String("input", "../data/es-latest.osm.pbf", "arquivo OSM PBF do Espírito Santo")
	saida := flag.String("output", "../data/es-road.graph.gz", "grafo viário compilado")
	flag.Parse()

	ctx, cancelar := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelar()
	log.Printf("importando vias de %s", *entrada)
	g, relatorio, err := osmimport.Construir(ctx, *entrada)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(
		"PBF processado: %d vias, %d vértices, %d arestas, %d segmentos ignorados em %s",
		relatorio.ViasAceitas,
		relatorio.Vertices,
		relatorio.Arestas,
		relatorio.SegmentosIgnorados,
		relatorio.Duracao,
	)
	if err := g.SalvarArquivo(*saida); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("grafo salvo em %s\n", *saida)
}
