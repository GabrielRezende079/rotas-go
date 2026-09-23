package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/storage"
)

func TestCriarListarExcluirBaseUsaStore(t *testing.T) {
	store := &storeFake{}
	service := NewBaseService(store)

	base, err := service.CriarBase(context.Background(), dtos.CriarBaseRequest{
		Nome: "Depósito Centro",
		Lat:  -20.3155,
		Lng:  -40.3128,
	})
	if err != nil {
		t.Fatal(err)
	}
	if base.ID == 0 || base.Nome != "Depósito Centro" {
		t.Fatalf("base criada incorreta: %+v", base)
	}

	lista, err := service.ListarBases(context.Background(), storage.ListaFiltro{Limite: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(lista.Items) != 1 || lista.Items[0].Nome != "Depósito Centro" || lista.Total != 1 {
		t.Fatalf("listagem de bases incorreta: %+v", lista)
	}

	if err := service.ExcluirBase(context.Background(), base.ID); err != nil {
		t.Fatalf("excluir base: %v", err)
	}
}

func TestCriarBaseValidaNomeECoordenadas(t *testing.T) {
	casos := []struct {
		nome   string
		store  BaseStorage
		req    dtos.CriarBaseRequest
		codigo int
	}{
		{nome: "sem store", store: nil, req: dtos.CriarBaseRequest{Nome: "B", Lat: -20, Lng: -40}, codigo: http.StatusServiceUnavailable},
		{nome: "sem nome", store: &storeFake{}, req: dtos.CriarBaseRequest{Lat: -20, Lng: -40}, codigo: http.StatusBadRequest},
		{nome: "latitude inválida", store: &storeFake{}, req: dtos.CriarBaseRequest{Nome: "B", Lat: -120, Lng: -40}, codigo: http.StatusBadRequest},
		{nome: "longitude inválida", store: &storeFake{}, req: dtos.CriarBaseRequest{Nome: "B", Lat: -20, Lng: -500}, codigo: http.StatusBadRequest},
	}
	for _, caso := range casos {
		service := NewBaseService(caso.store)
		_, err := service.CriarBase(context.Background(), caso.req)
		var httpErr ErroRequisicao
		if !errors.As(err, &httpErr) || httpErr.Codigo != caso.codigo {
			t.Fatalf("%s: esperava %d, obtido %v", caso.nome, caso.codigo, err)
		}
	}
}

func TestExcluirBaseInexistenteRetorna404(t *testing.T) {
	store := &storeFake{baseInexistente: true}
	service := NewBaseService(store)
	err := service.ExcluirBase(context.Background(), 99)
	var httpErr ErroRequisicao
	if !errors.As(err, &httpErr) || httpErr.Codigo != http.StatusNotFound {
		t.Fatalf("esperava 404, obtido %v", err)
	}
}

func TestListarBasesRetornaEnvelopePaginado(t *testing.T) {
	store := &storeFake{}
	service := NewBaseService(store)
	for i := 0; i < 3; i++ {
		if _, err := service.CriarBase(context.Background(), dtos.CriarBaseRequest{
			Nome: fmt.Sprintf("Depósito %d", i), Lat: -20, Lng: -40,
		}); err != nil {
			t.Fatal(err)
		}
	}
	page, err := service.ListarBases(context.Background(), storage.ListaFiltro{Termo: "Depósito", Limite: 7})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 3 || len(page.Items) != 3 || page.Limite != 7 || page.Offset != 0 {
		t.Fatalf("envelope paginado incorreto: %+v", page)
	}
	if page.Items[0].Nome != "Depósito 0" {
		t.Fatalf("ordem alfabética esperada, obtido %s", page.Items[0].Nome)
	}
}

func TestListarBasesNormalizaFiltro(t *testing.T) {
	service := NewBaseService(&storeFake{})
	page, err := service.ListarBases(context.Background(), storage.ListaFiltro{Limite: 999, Offset: -3})
	if err != nil {
		t.Fatal(err)
	}
	if page.Limite != 50 || page.Offset != 0 {
		t.Fatalf("filtro não normalizado: limite=%d offset=%d", page.Limite, page.Offset)
	}
}

func TestListarBasesSemStoreRetorna503(t *testing.T) {
	service := NewBaseService(nil)
	_, err := service.ListarBases(context.Background(), storage.ListaFiltro{Limite: 7})
	var httpErr ErroRequisicao
	if !errors.As(err, &httpErr) || httpErr.Codigo != http.StatusServiceUnavailable {
		t.Fatalf("esperava 503, obtido %v", err)
	}
}
