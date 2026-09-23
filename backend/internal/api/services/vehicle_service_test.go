package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"rotas-go/internal/api/dtos"
)

func TestCriarListarExcluirVeiculoUsaStore(t *testing.T) {
	store := &storeFake{}
	service := NewVehicleService(store)

	v, err := service.CriarVeiculo(context.Background(), dtos.CriarVeiculoRequest{
		Modelo:       "Fiorino",
		Categoria:    "Furgão",
		Placa:        "abc-1d23",
		Status:       "Disponível",
		Kilometragem: 125000,
		Velocidade:   0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if v.ID == 0 || v.Modelo != "Fiorino" || v.Placa != "ABC1D23" {
		t.Fatalf("veículo criado incorreto: %+v", v)
	}

	lista, err := service.ListarVeiculos(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(lista) != 1 || lista[0].Placa != "ABC1D23" {
		t.Fatalf("listagem de veículos incorreta: %+v", lista)
	}

	if err := service.ExcluirVeiculo(context.Background(), v.ID); err != nil {
		t.Fatalf("excluir veículo: %v", err)
	}
}

func TestCriarVeiculoPlacaDuplicadaRetorna409(t *testing.T) {
	store := &storeFake{placaDuplicada: true}
	service := NewVehicleService(store)
	_, err := service.CriarVeiculo(context.Background(), dtos.CriarVeiculoRequest{
		Modelo: "Fiorino", Categoria: "Furgão", Placa: "ABC1D23", Status: "Disponível",
	})
	var httpErr ErroRequisicao
	if !errors.As(err, &httpErr) || httpErr.Codigo != http.StatusConflict {
		t.Fatalf("esperava 409, obtido %v", err)
	}
}

func TestExcluirVeiculoInexistenteRetorna404(t *testing.T) {
	store := &storeFake{veiculoInexistente: true}
	service := NewVehicleService(store)
	err := service.ExcluirVeiculo(context.Background(), 99)
	var httpErr ErroRequisicao
	if !errors.As(err, &httpErr) || httpErr.Codigo != http.StatusNotFound {
		t.Fatalf("esperava 404, obtido %v", err)
	}
}

func TestCriarVeiculoValidaCampos(t *testing.T) {
	valido := dtos.CriarVeiculoRequest{Modelo: "Uno", Categoria: "Carro", Placa: "ABC1234", Status: "Disponível"}
	casos := []struct {
		nome   string
		store  VehicleStorage
		req    dtos.CriarVeiculoRequest
		codigo int
	}{
		{nome: "sem store", store: nil, req: valido, codigo: http.StatusServiceUnavailable},
		{nome: "sem modelo", store: &storeFake{}, req: dtos.CriarVeiculoRequest{Categoria: "Carro", Placa: "ABC1234", Status: "Disponível"}, codigo: http.StatusBadRequest},
		{nome: "categoria inválida", store: &storeFake{}, req: dtos.CriarVeiculoRequest{Modelo: "Uno", Categoria: "Nave", Placa: "ABC1234", Status: "Disponível"}, codigo: http.StatusBadRequest},
		{nome: "placa vazia", store: &storeFake{}, req: dtos.CriarVeiculoRequest{Modelo: "Uno", Categoria: "Carro", Placa: " ", Status: "Disponível"}, codigo: http.StatusBadRequest},
		{nome: "placa inválida", store: &storeFake{}, req: dtos.CriarVeiculoRequest{Modelo: "Uno", Categoria: "Carro", Placa: "12AB34", Status: "Disponível"}, codigo: http.StatusBadRequest},
		{nome: "status inválido", store: &storeFake{}, req: dtos.CriarVeiculoRequest{Modelo: "Uno", Categoria: "Carro", Placa: "ABC1234", Status: "Parado"}, codigo: http.StatusBadRequest},
		{nome: "kilometragem negativa", store: &storeFake{}, req: dtos.CriarVeiculoRequest{Modelo: "Uno", Categoria: "Carro", Placa: "ABC1234", Status: "Disponível", Kilometragem: -1}, codigo: http.StatusBadRequest},
		{nome: "velocidade negativa", store: &storeFake{}, req: dtos.CriarVeiculoRequest{Modelo: "Uno", Categoria: "Carro", Placa: "ABC1234", Status: "Disponível", Velocidade: -5}, codigo: http.StatusBadRequest},
	}
	for _, caso := range casos {
		service := NewVehicleService(caso.store)
		_, err := service.CriarVeiculo(context.Background(), caso.req)
		var httpErr ErroRequisicao
		if !errors.As(err, &httpErr) || httpErr.Codigo != caso.codigo {
			t.Fatalf("%s: esperava %d, obtido %v", caso.nome, caso.codigo, err)
		}
	}
}

func TestPlacaValidaFormatos(t *testing.T) {
	casos := []struct {
		placa string
		ok    bool
	}{
		{"ABC1D23", true},
		{"abc1d23", true},
		{"ABC-1234", true},
		{"abc 1234", true},
		{"ABC1234", true},
		{"ABC-1D23", true},
		{"ABC12345", false},
		{"ABC123", false},
		{"12AB34", false},
		{"ABCDEFG", false},
		{"1234567", false},
	}
	for _, caso := range casos {
		if placaValida(normalizarPlaca(caso.placa)) != caso.ok {
			t.Fatalf("placaValida(%q) = %v, esperado %v", caso.placa, !caso.ok, caso.ok)
		}
	}
}
