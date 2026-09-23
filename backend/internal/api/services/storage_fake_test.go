package services

import (
	"context"
	"time"

	"rotas-go/internal/grafo"
	"rotas-go/internal/storage"
)

func grafoDeTeste() *grafo.Grafo {
	g := grafo.NovoGrafo()
	g.AddVertice(grafo.Vertice{ID: 1, Lat: -20.30, Lng: -40.30})
	g.AddVertice(grafo.Vertice{ID: 2, Lat: -20.31, Lng: -40.31})
	g.AddVertice(grafo.Vertice{ID: 3, Lat: -20.32, Lng: -40.32})
	g.AddVertice(grafo.Vertice{ID: 4, Lat: -20.33, Lng: -40.33})
	g.AddArestaBidirecional(1, 2, 1.5, 2)
	g.AddArestaBidirecional(2, 3, 1.7, 3)
	g.AddArestaBidirecional(3, 4, 0.8, 1)
	g.SetMetadata("OSM teste", time.Unix(1, 0))
	return g
}

func grafoComAlternativas() *grafo.Grafo {
	g := grafo.NovoGrafo()
	for id := 1; id <= 10; id++ {
		g.AddVertice(grafo.Vertice{ID: int64(id), Lat: -20.0 - float64(id)/1000, Lng: -40.0 - float64(id)/1000})
	}
	// Corredor 1 (mais rápido): 1-2-3-4-5-6-10
	for i := 1; i <= 5; i++ {
		g.AddArestaBidirecional(int64(i), int64(i+1), 1, 2)
	}
	g.AddArestaBidirecional(6, 10, 1, 2)
	// Corredor 2 (mais lento): 1-7-8-9-10
	for _, par := range [][2]int64{{1, 7}, {7, 8}, {8, 9}, {9, 10}} {
		g.AddArestaBidirecional(par[0], par[1], 10, 4)
	}
	g.SetMetadata("OSM teste", time.Unix(1, 0))
	return g
}

// storeFake simula a persistência das rotas salvas, bases e veículos
// sem depender de um banco real.
type storeFake struct {
	ultimoNome         string
	ultimoAlgoritmo    string
	bases              []storage.Base
	proximaBaseID      int64
	baseInexistente    bool
	veiculos           []storage.Veiculo
	proximoVeiculo     int64
	veiculoInexistente bool
	placaDuplicada     bool
}

func (f *storeFake) Salvar(_ context.Context, nome, algoritmo string, veiculos int, distanciaKm, duracaoMin float64, _ []byte) (storage.RotaSalva, error) {
	f.ultimoNome = nome
	f.ultimoAlgoritmo = algoritmo
	return storage.RotaSalva{
		ID: 7, Nome: nome, Algoritmo: algoritmo,
		Veiculos: veiculos, DistanciaKm: distanciaKm, DuracaoMin: duracaoMin,
	}, nil
}

func (f *storeFake) Listar(_ context.Context, _ storage.ListaFiltro) ([]storage.RotaSalva, int, error) {
	return nil, 0, nil
}

func (f *storeFake) Buscar(_ context.Context, _ int64) (storage.RotaSalva, error) {
	return storage.RotaSalva{}, nil
}

func (f *storeFake) Excluir(_ context.Context, _ int64) error { return nil }

func (f *storeFake) CriarBase(_ context.Context, nome string, lat, lng float64) (storage.Base, error) {
	f.proximaBaseID++
	b := storage.Base{
		ID: f.proximaBaseID, Nome: nome, Lat: lat, Lng: lng,
		CriadaEm: time.Unix(1, 0).UTC(),
	}
	f.bases = append(f.bases, b)
	return b, nil
}

func (f *storeFake) ListarBases(_ context.Context, _ storage.ListaFiltro) ([]storage.Base, int, error) {
	return f.bases, len(f.bases), nil
}

func (f *storeFake) ExcluirBase(_ context.Context, id int64) error {
	if f.baseInexistente {
		return storage.ErrBaseNaoEncontrada
	}
	return nil
}

func (f *storeFake) CriarVeiculo(_ context.Context, modelo, categoria, placa, status string, kilometragem, velocidade float64) (storage.Veiculo, error) {
	if f.placaDuplicada {
		return storage.Veiculo{}, storage.ErrPlacaDuplicada
	}
	f.proximoVeiculo++
	v := storage.Veiculo{
		ID: f.proximoVeiculo, Modelo: modelo, Categoria: categoria, Placa: placa,
		Status: status, Kilometragem: kilometragem, Velocidade: velocidade,
		CriadoEm: time.Unix(1, 0).UTC(),
	}
	f.veiculos = append(f.veiculos, v)
	return v, nil
}

func (f *storeFake) ListarVeiculos(_ context.Context) ([]storage.Veiculo, error) {
	return f.veiculos, nil
}

func (f *storeFake) ExcluirVeiculo(_ context.Context, _ int64) error {
	if f.veiculoInexistente {
		return storage.ErrVeiculoNaoEncontrado
	}
	return nil
}
