package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/storage"
)

func TestSalvarRotaSemPersistenciaRetorna503(t *testing.T) {
	service := NewSavedRouteService(nil)
	_, err := service.SalvarRota(context.Background(), dtos.SalvarRotaRequest{
		Nome: "teste", Algoritmo: "dijkstra",
	})
	var httpErr ErroRequisicao
	if err == nil || !errors.As(err, &httpErr) || httpErr.Codigo != http.StatusServiceUnavailable {
		t.Fatalf("esperava 503 sem persistência, obtido %v", err)
	}
}

func TestSalvarRotaUsaStoreFornecido(t *testing.T) {
	store := &storeFake{}
	service := NewSavedRouteService(store)
	resumo, err := service.SalvarRota(context.Background(), dtos.SalvarRotaRequest{
		Nome:      "entrega norte",
		Algoritmo: "astar",
		Rotas: []dtos.LoteItemResponse{
			{ID: "v1", RotaResponse: dtos.RotaResponse{DistanciaKm: 10, DuracaoEstimadaMin: 15}},
			{ID: "v2", RotaResponse: dtos.RotaResponse{DistanciaKm: 20, DuracaoEstimadaMin: 25}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resumo.Veiculos != 2 || resumo.DistanciaKm != 30 || resumo.DuracaoMin != 40 {
		t.Fatalf("resumo incorreto: %+v", resumo)
	}
	if store.ultimoNome != "entrega norte" || store.ultimoAlgoritmo != "astar" {
		t.Fatalf("store chamado com valores errados: %q %q", store.ultimoNome, store.ultimoAlgoritmo)
	}
}

func TestListarRotasSalvasRetornaEnvelopePaginado(t *testing.T) {
	service := NewSavedRouteService(&storeFake{})
	page, err := service.ListarRotasSalvas(context.Background(), storage.ListaFiltro{Termo: "norte", Limite: 7})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 || len(page.Items) != 0 || page.Limite != 7 || page.Offset != 0 {
		t.Fatalf("envelope paginado incorreto: %+v", page)
	}
}

func TestListarRotasSalvasSemStoreRetorna503(t *testing.T) {
	service := NewSavedRouteService(nil)
	_, err := service.ListarRotasSalvas(context.Background(), storage.ListaFiltro{Limite: 7})
	var httpErr ErroRequisicao
	if !errors.As(err, &httpErr) || httpErr.Codigo != http.StatusServiceUnavailable {
		t.Fatalf("esperava 503, obtido %v", err)
	}
}
