package services

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/storage"
)

// BaseStorage é a persistência de bases (localizações padrão).
type BaseStorage interface {
	CriarBase(ctx context.Context, nome string, lat, lng float64) (storage.Base, error)
	ListarBases(ctx context.Context, f storage.ListaFiltro) ([]storage.Base, int, error)
	ExcluirBase(ctx context.Context, id int64) error
}

// BaseService valida e consulta as bases cadastradas.
type BaseService struct {
	store BaseStorage
}

// NewBaseService cria o serviço de bases.
// O store pode ser nil (DATABASE_URL ausente) e nesse caso responde 503.
func NewBaseService(store BaseStorage) *BaseService {
	return &BaseService{store: store}
}

// CriarBase valida e persiste uma base (localização padrão).
func (s *BaseService) CriarBase(ctx context.Context, req dtos.CriarBaseRequest) (dtos.BaseResponse, error) {
	if s.store == nil {
		return dtos.BaseResponse{}, semPersistencia()
	}
	nome := strings.TrimSpace(req.Nome)
	if nome == "" {
		return dtos.BaseResponse{}, erroHTTP(http.StatusBadRequest, "informe um nome para a base")
	}
	if req.Lat < -90 || req.Lat > 90 {
		return dtos.BaseResponse{}, erroHTTP(http.StatusBadRequest, "latitude fora do intervalo [-90, 90]")
	}
	if req.Lng < -180 || req.Lng > 180 {
		return dtos.BaseResponse{}, erroHTTP(http.StatusBadRequest, "longitude fora do intervalo [-180, 180]")
	}
	b, err := s.store.CriarBase(ctx, nome, req.Lat, req.Lng)
	if err != nil {
		return dtos.BaseResponse{}, err
	}
	return resumoDeBase(b), nil
}

// ListarBases devolve uma página de bases, com filtragem por nome.
func (s *BaseService) ListarBases(ctx context.Context, f storage.ListaFiltro) (dtos.BasesResponse, error) {
	if s.store == nil {
		return dtos.BasesResponse{}, semPersistencia()
	}
	f = normalizarFiltro(f)
	lista, total, err := s.store.ListarBases(ctx, f)
	if err != nil {
		return dtos.BasesResponse{}, err
	}
	resumo := make([]dtos.BaseResponse, 0, len(lista))
	for _, b := range lista {
		resumo = append(resumo, resumoDeBase(b))
	}
	return dtos.BasesResponse{Items: resumo, Total: total, Limite: f.Limite, Offset: f.Offset}, nil
}

// ExcluirBase remove uma base.
func (s *BaseService) ExcluirBase(ctx context.Context, id int64) error {
	if s.store == nil {
		return semPersistencia()
	}
	err := s.store.ExcluirBase(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrBaseNaoEncontrada) {
			return erroHTTP(http.StatusNotFound, "base não encontrada")
		}
		return err
	}
	return nil
}

func resumoDeBase(b storage.Base) dtos.BaseResponse {
	return dtos.BaseResponse{
		ID:       b.ID,
		Nome:     b.Nome,
		Lat:      b.Lat,
		Lng:      b.Lng,
		CriadaEm: b.CriadaEm,
	}
}
