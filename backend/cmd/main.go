package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"rotas-go/internal/grafo"
	"rotas-go/routes"
)

func main() {
	// O grafo é construído fora do pacote de rotas e injetado no serviço.
	grafoES := grafo.NovoGrafoES()
	service := routes.NewRouteService(grafoES)
	controller := routes.NewRouteController(service)

	engine := gin.Default()
	engine.Use(routes.CORSMiddleware())

	api := engine.Group("/api/v1")
	api.GET("/nodes", controller.ListarNos)
	api.POST("/route", controller.CalcularRota)

	if err := engine.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
