package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// RotaSalva é o registro persistido de uma rota calculada.
// Dados guarda o snapshot JSON dos veículos (origem/destino/geometria).
type RotaSalva struct {
	ID          int64           `json:"id"`
	Nome        string          `json:"name"`
	Algoritmo   string          `json:"algorithm"`
	CriadaEm    time.Time       `json:"created_at"`
	Veiculos    int             `json:"vehicle_count"`
	DistanciaKm float64         `json:"total_distance_km"`
	DuracaoMin  float64         `json:"total_duration_minutes"`
	Dados       json.RawMessage `json:"-"`
}

// Salvar insere uma rota e devolve o registro criado.
func (s *PostgresStore) Salvar(ctx context.Context, nome, algoritmo string, veiculos int, distanciaKm, duracaoMin float64, dados []byte) (RotaSalva, error) {
	var rs RotaSalva
	err := s.pool.QueryRow(ctx, `
		INSERT INTO saved_routes (name, algorithm, vehicle_count, total_distance_km, total_duration_minutes, data)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		RETURNING id, name, algorithm, created_at, vehicle_count, total_distance_km, total_duration_minutes`,
		nome, algoritmo, veiculos, distanciaKm, duracaoMin, dados,
	).Scan(&rs.ID, &rs.Nome, &rs.Algoritmo, &rs.CriadaEm, &rs.Veiculos, &rs.DistanciaKm, &rs.DuracaoMin)
	return rs, err
}

// Listar devolve uma página de resumos das rotas salvas,
// da mais recente para a mais antiga. O termo filtra por nome;
// o total é o nº de registros que casam com o filtro.
func (s *PostgresStore) Listar(ctx context.Context, f ListaFiltro) ([]RotaSalva, int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, algorithm, created_at, vehicle_count, total_distance_km, total_duration_minutes, count(*) OVER()::int AS total
		FROM saved_routes
		WHERE name ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		f.Termo, f.Limite, f.Offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("listar rotas salvas: %w", err)
	}
	defer rows.Close()

	lista := make([]RotaSalva, 0)
	total := 0
	for rows.Next() {
		var rs RotaSalva
		if err := rows.Scan(&rs.ID, &rs.Nome, &rs.Algoritmo, &rs.CriadaEm, &rs.Veiculos, &rs.DistanciaKm, &rs.DuracaoMin, &total); err != nil {
			return nil, 0, fmt.Errorf("ler rota salva: %w", err)
		}
		lista = append(lista, rs)
	}
	return lista, total, rows.Err()
}

// Buscar devolve a rota salva com o snapshot completo.
func (s *PostgresStore) Buscar(ctx context.Context, id int64) (RotaSalva, error) {
	var rs RotaSalva
	var dados []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, algorithm, created_at, vehicle_count, total_distance_km, total_duration_minutes, data
		FROM saved_routes
		WHERE id = $1`, id,
	).Scan(&rs.ID, &rs.Nome, &rs.Algoritmo, &rs.CriadaEm, &rs.Veiculos, &rs.DistanciaKm, &rs.DuracaoMin, &dados)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RotaSalva{}, ErrNaoEncontrada
		}
		return RotaSalva{}, fmt.Errorf("buscar rota salva %d: %w", id, err)
	}
	rs.Dados = dados
	return rs, nil
}

// Excluir remove uma rota salva.
func (s *PostgresStore) Excluir(ctx context.Context, id int64) error {
	res, err := s.pool.Exec(ctx, `DELETE FROM saved_routes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("excluir rota salva %d: %w", id, err)
	}
	if res.RowsAffected() == 0 {
		return ErrNaoEncontrada
	}
	return nil
}
