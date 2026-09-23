package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/grafo"
	"rotas-go/internal/storage"
	"rotas-go/routes"
)

func main() {
	g := carregarGrafo()
	meta := g.Metadata()
	log.Printf(
		"grafo carregado: fonte=%q vertices=%d arestas=%d gerado_em=%s",
		meta.Fonte,
		meta.Vertices,
		meta.Arestas,
		meta.GeradoEm.Format(time.RFC3339),
	)

	var store routes.SavedRouteStore
	if url := os.Getenv("DATABASE_URL"); url != "" {
		pg := inicializarArmazenamento(url)
		defer pg.Close()
		store = pg
	} else {
		log.Println("DATABASE_URL não definida; rotas salvas desabilitadas")
	}

	service := routes.NewRouteService(g, store)
	controller := routes.NewRouteController(service)
	engine := gin.Default()
	_ = engine.SetTrustedProxies(nil)
	engine.Use(routes.CORSMiddleware())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	api := engine.Group("/api/v1")
	api.GET("/info", controller.Info)
	api.POST("/route", controller.CalcularRota)
	api.POST("/route/alternatives", controller.CalcularAlternativas)
	api.POST("/routes", controller.CalcularLote)
	api.POST("/routes/saved", controller.SalvarRota)
	api.GET("/routes/saved", controller.ListarRotasSalvas)
	api.GET("/routes/saved/:id", controller.BuscarRotaSalva)
	api.DELETE("/routes/saved/:id", controller.ExcluirRotaSalva)
	api.POST("/bases", controller.CriarBase)
	api.GET("/bases", controller.ListarBases)
	api.DELETE("/bases/:id", controller.ExcluirBase)

	if err := engine.Run(":" + porta()); err != nil {
		log.Fatal(err)
	}
}

func inicializarArmazenamento(url string) *storage.PostgresStore {
	ctx := context.Background()
	pg, err := storage.NovoPostgresStore(ctx, url)
	if err != nil {
		log.Fatalf("DATABASE_URL inválida: %v", err)
	}
	if err := pg.Migrar(ctx); err != nil {
		pg.Close()
		log.Fatalf("migração do banco falhou: %v", err)
	}
	log.Println("persistência PostgreSQL habilitada (rotas salvas)")
	return pg
}

func porta() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}

func carregarGrafo() *grafo.Grafo {
	caminho := os.Getenv("ROTAS_GRAPH_FILE")
	if caminho == "" {
		caminho = filepath.Clean("../data/es-road.graph.gz")
	}
	if g, err := grafo.CarregarArquivo(caminho); err == nil {
		return g
	} else if !os.IsNotExist(err) {
		log.Printf("não foi possível carregar %s: %v", caminho, err)
	}
	log.Printf("ATENÇÃO: %s não existe; usando grafo didático", caminho)
	return grafo.NovoGrafoESDidatico()
}
