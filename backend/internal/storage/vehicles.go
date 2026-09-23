package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// Veiculo é um veículo da frota cadastrado.
type Veiculo struct {
	ID           int64     `json:"id"`
	Modelo       string    `json:"modelo"`
	Categoria    string    `json:"categoria"`
	Placa        string    `json:"placa"`
	Status       string    `json:"status"`
	Kilometragem float64   `json:"kilometragem"`
	Velocidade   float64   `json:"velocidade"`
	CriadoEm     time.Time `json:"created_at"`
}

// CriarVeiculo insere um veículo e devolve o registro criado.
// Placa deve estar normalizada (maiúscula, sem separadores).
func (s *PostgresStore) CriarVeiculo(ctx context.Context, modelo, categoria, placa, status string, kilometragem, velocidade float64) (Veiculo, error) {
	var v Veiculo
	err := s.pool.QueryRow(ctx, `
		INSERT INTO vehicles (modelo, categoria, placa, status, kilometragem, velocidade)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, modelo, categoria, placa, status, kilometragem, velocidade, created_at`,
		modelo, categoria, placa, status, kilometragem, velocidade,
	).Scan(&v.ID, &v.Modelo, &v.Categoria, &v.Placa, &v.Status, &v.Kilometragem, &v.Velocidade, &v.CriadoEm)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Veiculo{}, ErrPlacaDuplicada
		}
		return Veiculo{}, fmt.Errorf("criar veículo: %w", err)
	}
	return v, nil
}

// ListarVeiculos devolve os veículos da frota, do mais recente para o mais antigo.
func (s *PostgresStore) ListarVeiculos(ctx context.Context) ([]Veiculo, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, modelo, categoria, placa, status, kilometragem, velocidade, created_at
		FROM vehicles
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("listar veículos: %w", err)
	}
	defer rows.Close()

	lista := make([]Veiculo, 0)
	for rows.Next() {
		var v Veiculo
		if err := rows.Scan(&v.ID, &v.Modelo, &v.Categoria, &v.Placa, &v.Status, &v.Kilometragem, &v.Velocidade, &v.CriadoEm); err != nil {
			return nil, fmt.Errorf("ler veículo: %w", err)
		}
		lista = append(lista, v)
	}
	return lista, rows.Err()
}

// ExcluirVeiculo remove um veículo da frota.
func (s *PostgresStore) ExcluirVeiculo(ctx context.Context, id int64) error {
	res, err := s.pool.Exec(ctx, `DELETE FROM vehicles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("excluir veículo %d: %w", id, err)
	}
	if res.RowsAffected() == 0 {
		return ErrVeiculoNaoEncontrado
	}
	return nil
}
