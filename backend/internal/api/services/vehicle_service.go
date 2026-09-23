package services

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"rotas-go/internal/api/dtos"
	"rotas-go/internal/storage"
)

// categoriasValidas e statusValidos definem os valores aceitos para a frota.
var categoriasValidas = map[string]bool{
	"Carro":       true,
	"Caminhão":    true,
	"Caminhonete": true,
	"Furgão":      true,
}

var statusValidos = map[string]bool{
	"Disponível":    true,
	"Indisponível":  true,
	"Em uso":        true,
	"Em Manutenção": true,
}

// VehicleStorage é a persistência de veículos da frota.
type VehicleStorage interface {
	CriarVeiculo(ctx context.Context, modelo, categoria, placa, status string, kilometragem, velocidade float64) (storage.Veiculo, error)
	ListarVeiculos(ctx context.Context) ([]storage.Veiculo, error)
	ExcluirVeiculo(ctx context.Context, id int64) error
}

// VehicleService valida e consulta os veículos da frota.
type VehicleService struct {
	store VehicleStorage
}

// NewVehicleService cria o serviço de veículos.
// O store pode ser nil (DATABASE_URL ausente) e nesse caso responde 503.
func NewVehicleService(store VehicleStorage) *VehicleService {
	return &VehicleService{store: store}
}

// CriarVeiculo valida e persiste um veículo da frota.
func (s *VehicleService) CriarVeiculo(ctx context.Context, req dtos.CriarVeiculoRequest) (dtos.VeiculoResponse, error) {
	if s.store == nil {
		return dtos.VeiculoResponse{}, semPersistencia()
	}
	modelo := strings.TrimSpace(req.Modelo)
	if modelo == "" {
		return dtos.VeiculoResponse{}, erroHTTP(http.StatusBadRequest, "informe o modelo do veículo")
	}
	if !categoriasValidas[req.Categoria] {
		return dtos.VeiculoResponse{}, erroHTTP(http.StatusBadRequest, "categoria inválida")
	}
	placa := normalizarPlaca(req.Placa)
	if placa == "" {
		return dtos.VeiculoResponse{}, erroHTTP(http.StatusBadRequest, "informe a placa do veículo")
	}
	if !placaValida(placa) {
		return dtos.VeiculoResponse{}, erroHTTP(http.StatusBadRequest, "placa inválida (use ABC1D23 ou ABC-1234)")
	}
	if !statusValidos[req.Status] {
		return dtos.VeiculoResponse{}, erroHTTP(http.StatusBadRequest, "status inválido")
	}
	if req.Kilometragem < 0 {
		return dtos.VeiculoResponse{}, erroHTTP(http.StatusBadRequest, "kilometragem não pode ser negativa")
	}
	if req.Velocidade < 0 {
		return dtos.VeiculoResponse{}, erroHTTP(http.StatusBadRequest, "velocidade não pode ser negativa")
	}
	v, err := s.store.CriarVeiculo(ctx, modelo, req.Categoria, placa, req.Status, req.Kilometragem, req.Velocidade)
	if err != nil {
		if errors.Is(err, storage.ErrPlacaDuplicada) {
			return dtos.VeiculoResponse{}, erroHTTP(http.StatusConflict, "placa já cadastrada")
		}
		return dtos.VeiculoResponse{}, err
	}
	return resumoVeiculo(v), nil
}

// ListarVeiculos devolve todos os veículos da frota.
func (s *VehicleService) ListarVeiculos(ctx context.Context) ([]dtos.VeiculoResponse, error) {
	if s.store == nil {
		return nil, semPersistencia()
	}
	lista, err := s.store.ListarVeiculos(ctx)
	if err != nil {
		return nil, err
	}
	resumo := make([]dtos.VeiculoResponse, 0, len(lista))
	for _, v := range lista {
		resumo = append(resumo, resumoVeiculo(v))
	}
	return resumo, nil
}

// ExcluirVeiculo remove um veículo da frota.
func (s *VehicleService) ExcluirVeiculo(ctx context.Context, id int64) error {
	if s.store == nil {
		return semPersistencia()
	}
	err := s.store.ExcluirVeiculo(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrVeiculoNaoEncontrado) {
			return erroHTTP(http.StatusNotFound, "veículo não encontrado")
		}
		return err
	}
	return nil
}

// normalizarPlaca devolve a placa em maiúsculas e sem separadores.
func normalizarPlaca(placa string) string {
	placa = strings.ToUpper(strings.TrimSpace(placa))
	placa = strings.ReplaceAll(placa, "-", "")
	placa = strings.ReplaceAll(placa, " ", "")
	return placa
}

// placaValida aceita o formato antigo (ABC1234) e o Mercosul (ABC1D23).
func placaValida(placa string) bool {
	placa = normalizarPlaca(placa)
	if len(placa) != 7 {
		return false
	}
	for i := 0; i < 3; i++ {
		if placa[i] < 'A' || placa[i] > 'Z' {
			return false
		}
	}
	if placa[3] < '0' || placa[3] > '9' {
		return false
	}
	if placa[4] >= 'A' && placa[4] <= 'Z' {
		return (placa[5] >= '0' && placa[5] <= '9') && (placa[6] >= '0' && placa[6] <= '9')
	}
	return (placa[4] >= '0' && placa[4] <= '9') && (placa[5] >= '0' && placa[5] <= '9') && (placa[6] >= '0' && placa[6] <= '9')
}

func resumoVeiculo(v storage.Veiculo) dtos.VeiculoResponse {
	return dtos.VeiculoResponse{
		ID:           v.ID,
		Modelo:       v.Modelo,
		Categoria:    v.Categoria,
		Placa:        v.Placa,
		Status:       v.Status,
		Kilometragem: v.Kilometragem,
		Velocidade:   v.Velocidade,
		CriadoEm:     v.CriadoEm,
	}
}
