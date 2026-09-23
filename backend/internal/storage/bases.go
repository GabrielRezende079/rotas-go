package storage

import (
	"context"
	"fmt"
	"time"
)

// Base é uma localização padrão (depósito ou ponto de apoio) persistida.
type Base struct {
	ID       int64     `json:"id"`
	Nome     string    `json:"name"`
	Lat      float64   `json:"lat"`
	Lng      float64   `json:"lng"`
	CriadaEm time.Time `json:"created_at"`
}

// CriarBase insere uma base e devolve o registro criado.
func (s *PostgresStore) CriarBase(ctx context.Context, nome string, lat, lng float64) (Base, error) {
	var b Base
	err := s.pool.QueryRow(ctx, `
		INSERT INTO bases (name, lat, lng)
		VALUES ($1, $2, $3)
		RETURNING id, name, lat, lng, created_at`,
		nome, lat, lng,
	).Scan(&b.ID, &b.Nome, &b.Lat, &b.Lng, &b.CriadaEm)
	return b, err
}

// ListarBases devolve as bases em ordem alfabética pelo nome.
func (s *PostgresStore) ListarBases(ctx context.Context) ([]Base, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, lat, lng, created_at
		FROM bases
		ORDER BY name ASC`)
	if err != nil {
		return nil, fmt.Errorf("listar bases: %w", err)
	}
	defer rows.Close()

	lista := make([]Base, 0)
	for rows.Next() {
		var b Base
		if err := rows.Scan(&b.ID, &b.Nome, &b.Lat, &b.Lng, &b.CriadaEm); err != nil {
			return nil, fmt.Errorf("ler base: %w", err)
		}
		lista = append(lista, b)
	}
	return lista, rows.Err()
}

// ExcluirBase remove uma base.
func (s *PostgresStore) ExcluirBase(ctx context.Context, id int64) error {
	res, err := s.pool.Exec(ctx, `DELETE FROM bases WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("excluir base %d: %w", id, err)
	}
	if res.RowsAffected() == 0 {
		return ErrBaseNaoEncontrada
	}
	return nil
}
