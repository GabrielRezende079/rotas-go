package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/api/controllers"
	"rotas-go/internal/api/services"
	"rotas-go/internal/grafo"
	"rotas-go/internal/storage"
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

	var store *storage.PostgresStore
	if url := os.Getenv("DATABASE_URL"); url != "" {
		store = inicializarArmazenamento(url)
		defer store.Close()
	} else {
		log.Println("DATABASE_URL não definida; rotas salvas, bases e veículos desabilitados")
	}

	infoService := services.NewInfoService(g)
	routeService := services.NewRouteService(g)
	alternativeService := services.NewAlternativeService(g)
	savedRouteService := services.NewSavedRouteService(store)
	baseService := services.NewBaseService(store)
	vehicleService := services.NewVehicleService(store)

	infoController := controllers.NewInfoController(infoService)
	routeController := controllers.NewRouteController(routeService)
	alternativeController := controllers.NewAlternativeController(alternativeService)
	savedRouteController := controllers.NewSavedRouteController(savedRouteService)
	baseController := controllers.NewBaseController(baseService)
	vehicleController := controllers.NewVehicleController(vehicleService)

	engine := gin.Default()
	_ = engine.SetTrustedProxies(nil)
	engine.Use(controllers.CORSMiddleware())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	api := engine.Group("/api/v1")
	api.GET("/info", infoController.Info)
	api.POST("/route", routeController.CalcularRota)
	api.POST("/route/alternatives", alternativeController.CalcularAlternativas)
	api.POST("/routes", routeController.CalcularLote)
	api.POST("/routes/saved", savedRouteController.SalvarRota)
	api.GET("/routes/saved", savedRouteController.ListarRotasSalvas)
	api.GET("/routes/saved/:id", savedRouteController.BuscarRotaSalva)
	api.DELETE("/routes/saved/:id", savedRouteController.ExcluirRotaSalva)
	api.POST("/bases", baseController.CriarBase)
	api.GET("/bases", baseController.ListarBases)
	api.DELETE("/bases/:id", baseController.ExcluirBase)
	api.POST("/vehicles", vehicleController.CriarVeiculo)
	api.GET("/vehicles", vehicleController.ListarVeiculos)
	api.DELETE("/vehicles/:id", vehicleController.ExcluirVeiculo)

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
	log.Println("persistência PostgreSQL habilitada (rotas salvas, bases, veículos)")
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
