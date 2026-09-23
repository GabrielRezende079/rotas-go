// Package storage implementa a persistência das rotas salvas sobre PostgreSQL.
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"rotas-go/migrations"
)

// ErrNaoEncontrada indica que a rota salva não existe.
var ErrNaoEncontrada = errors.New("rota salva não encontrada")

// ErrBaseNaoEncontrada indica que a base não existe.
var ErrBaseNaoEncontrada = errors.New("base não encontrada")

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

// PostgresStore consulta e grava rotas salvas usando pgx.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NovoPostgresStore cria o pool de conexões e valida a conexão.
func NovoPostgresStore(ctx context.Context, url string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("criar pool de conexões: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping no banco: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

// Close encerra o pool de conexões.
func (s *PostgresStore) Close() { s.pool.Close() }

// Migrar aplica os arquivos SQL de backend/migrations em ordem alfabética.
func (s *PostgresStore) Migrar(ctx context.Context) error {
	entradas, err := migrations.FS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("ler migrações: %w", err)
	}
	nomes := make([]string, 0, len(entradas))
	for _, e := range entradas {
		if !e.IsDir() {
			nomes = append(nomes, e.Name())
		}
	}
	sort.Strings(nomes)
	for _, nome := range nomes {
		sql, err := migrations.FS.ReadFile(nome)
		if err != nil {
			return fmt.Errorf("ler migração %s: %w", nome, err)
		}
		if _, err := s.pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("aplicar migração %s: %w", nome, err)
		}
	}
	return nil
}

// Base é uma localização padrão (depósito ou ponto de apoio) persistida.
type Base struct {
	ID       int64     `json:"id"`
	Nome     string    `json:"name"`
	Lat      float64   `json:"lat"`
	Lng      float64   `json:"lng"`
	CriadaEm time.Time `json:"created_at"`
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

// Listar devolve os resumos das rotas salvas, da mais recente para a mais antiga.
func (s *PostgresStore) Listar(ctx context.Context) ([]RotaSalva, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, algorithm, created_at, vehicle_count, total_distance_km, total_duration_minutes
		FROM saved_routes
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("listar rotas salvas: %w", err)
	}
	defer rows.Close()

	lista := make([]RotaSalva, 0)
	for rows.Next() {
		var rs RotaSalva
		if err := rows.Scan(&rs.ID, &rs.Nome, &rs.Algoritmo, &rs.CriadaEm, &rs.Veiculos, &rs.DistanciaKm, &rs.DuracaoMin); err != nil {
			return nil, fmt.Errorf("ler rota salva: %w", err)
		}
		lista = append(lista, rs)
	}
	return lista, rows.Err()
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
