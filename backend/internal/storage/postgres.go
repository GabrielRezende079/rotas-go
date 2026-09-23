// Package storage implementa a persistência das rotas salvas sobre PostgreSQL.
package storage

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"rotas-go/migrations"
)

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
