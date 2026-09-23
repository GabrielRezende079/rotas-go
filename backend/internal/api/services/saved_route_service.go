package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/storage"
)

// SavedRouteStorage é a persistência de rotas calculadas.
// O serviço depende apenas desta interface; a implementação PostgreSQL
// vive em rotas-go/internal/storage.
type SavedRouteStorage interface {
	Salvar(ctx context.Context, nome, algoritmo string, veiculos int, distanciaKm, duracaoMin float64, dados []byte) (storage.RotaSalva, error)
	Listar(ctx context.Context, f storage.ListaFiltro) ([]storage.RotaSalva, int, error)
	Buscar(ctx context.Context, id int64) (storage.RotaSalva, error)
	Excluir(ctx context.Context, id int64) error
}

// SavedRouteService persiste e consulta rotas salvas.
type SavedRouteService struct {
	store SavedRouteStorage
}

// NewSavedRouteService cria o serviço de rotas salvas.
// O store pode ser nil (DATABASE_URL ausente) e nesse caso responde 503.
func NewSavedRouteService(store SavedRouteStorage) *SavedRouteService {
	return &SavedRouteService{store: store}
}

// SalvarRota persiste o snapshot calculado de um lote de rotas.
func (s *SavedRouteService) SalvarRota(ctx context.Context, req dtos.SalvarRotaRequest) (dtos.RotaSalvaSummary, error) {
	if s.store == nil {
		return dtos.RotaSalvaSummary{}, semPersistencia()
	}
	nome := strings.TrimSpace(req.Nome)
	if nome == "" {
		return dtos.RotaSalvaSummary{}, erroHTTP(http.StatusBadRequest, "informe um nome para a rota")
	}
	algoritmo := strings.ToLower(strings.TrimSpace(req.Algoritmo))
	if !algoritmoValido(algoritmo) {
		return dtos.RotaSalvaSummary{}, erroHTTP(http.StatusBadRequest, fmt.Sprintf("algoritmo desconhecido: %s", req.Algoritmo))
	}
	if len(req.Rotas) == 0 {
		return dtos.RotaSalvaSummary{}, erroHTTP(http.StatusBadRequest, "a rota não possui veículos")
	}

	var totalDistancia, totalDuracao float64
	for _, r := range req.Rotas {
		totalDistancia += r.DistanciaKm
		totalDuracao += r.DuracaoEstimadaMin
	}
	dados, err := json.Marshal(req.Rotas)
	if err != nil {
		return dtos.RotaSalvaSummary{}, err
	}
	salva, err := s.store.Salvar(ctx, nome, algoritmo, len(req.Rotas), totalDistancia, totalDuracao, dados)
	if err != nil {
		return dtos.RotaSalvaSummary{}, err
	}
	return resumoDeRotaSalva(salva), nil
}

// ListarRotasSalvas devolve uma página de resumos, com filtragem por nome.
func (s *SavedRouteService) ListarRotasSalvas(ctx context.Context, f storage.ListaFiltro) (dtos.RotasSalvasResponse, error) {
	if s.store == nil {
		return dtos.RotasSalvasResponse{}, semPersistencia()
	}
	f = normalizarFiltro(f)
	lista, total, err := s.store.Listar(ctx, f)
	if err != nil {
		return dtos.RotasSalvasResponse{}, err
	}
	resumo := make([]dtos.RotaSalvaSummary, 0, len(lista))
	for _, r := range lista {
		resumo = append(resumo, resumoDeRotaSalva(r))
	}
	return dtos.RotasSalvasResponse{Items: resumo, Total: total, Limite: f.Limite, Offset: f.Offset}, nil
}

// BuscarRotaSalva devolve o snapshot completo de uma rota salva.
func (s *SavedRouteService) BuscarRotaSalva(ctx context.Context, id int64) (dtos.RotaSalvaDetalhe, error) {
	if s.store == nil {
		return dtos.RotaSalvaDetalhe{}, semPersistencia()
	}
	salva, err := s.store.Buscar(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNaoEncontrada) {
			return dtos.RotaSalvaDetalhe{}, erroHTTP(http.StatusNotFound, "rota salva não encontrada")
		}
		return dtos.RotaSalvaDetalhe{}, err
	}
	return dtos.RotaSalvaDetalhe{RotaSalvaSummary: resumoDeRotaSalva(salva), Rotas: salva.Dados}, nil
}

// ExcluirRotaSalva remove uma rota salva.
func (s *SavedRouteService) ExcluirRotaSalva(ctx context.Context, id int64) error {
	if s.store == nil {
		return semPersistencia()
	}
	err := s.store.Excluir(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNaoEncontrada) {
			return erroHTTP(http.StatusNotFound, "rota salva não encontrada")
		}
		return err
	}
	return nil
}

func resumoDeRotaSalva(r storage.RotaSalva) dtos.RotaSalvaSummary {
	return dtos.RotaSalvaSummary{
		ID:          r.ID,
		Nome:        r.Nome,
		Algoritmo:   r.Algoritmo,
		CriadaEm:    r.CriadaEm,
		Veiculos:    r.Veiculos,
		DistanciaKm: r.DistanciaKm,
		DuracaoMin:  r.DuracaoMin,
	}
}
