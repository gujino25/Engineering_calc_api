package postgres

import (
	"context"
	"enginer/internal/domain"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SystemsRepo struct {
	pool *pgxpool.Pool
}

func NewSystemRepo(pool *pgxpool.Pool) *SystemsRepo {
	return &SystemsRepo{
		pool: pool,
	}
}

func (r *SystemsRepo) Create(ctx context.Context, system domain.System) error {

	const query = `INSERT INTO systems (id, project_id, name, medium, purpose, created_at) VALUES ($1,$2,$3,$4,$5,$6)`

	_, err := r.pool.Exec(ctx, query, system.ID, system.ProjectID, system.Name, string(system.Medium), system.Purpose, system.CreatedAt)
	if err != nil {
		var pgrErr *pgconn.PgError
		if errors.As(err, &pgrErr) {
			switch pgrErr.Code {
			case "23505":
				return domain.ErrSystemAlreadyExists
			case "23503":
				return domain.ErrProjectNotFound
			}
		}
		return fmt.Errorf("create system: %w", err)
	}
	return nil
}

func (r *SystemsRepo) GetByID(ctx context.Context, id string) (domain.System, error) {
	const query = `SELECT id, project_id, name, medium, purpose, created_at FROM systems WHERE id = $1 `
	var system domain.System

	var medium string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&system.ID,
		&system.ProjectID,
		&system.Name,
		&medium,
		&system.Purpose,
		&system.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.System{}, domain.ErrSystemNotFound
		}
		return domain.System{}, fmt.Errorf("get system: %w", err)
	}
	system.Medium = domain.Medium(medium)
	return system, nil
}

func (r *SystemsRepo) List(ctx context.Context) ([]domain.System, error) {
	const query = `SELECT id, project_id, name, medium, purpose, created_at FROM systems ORDER BY created_at, id`

	rows, err := r.pool.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("get all systems: %w", err)
	}
	defer rows.Close()
	var systems []domain.System
	for rows.Next() {
		var medium string
		var s domain.System
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Name, &medium, &s.Purpose, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan systems: %w", err)
		}
		s.Medium = domain.Medium(medium)
		systems = append(systems, s)
	}
	return systems, rows.Err()
}

func (r *SystemsRepo) ListByProject(ctx context.Context, project_id string) ([]domain.System, error) {
	const query = `SELECT id, project_id,name,medium,purpose,created_at FROM systems WHERE project_id =$1 ORDER BY created_at, id`
	rows, err := r.pool.Query(ctx, query, project_id)
	if err != nil {
		return nil, fmt.Errorf("get all systems by project: %w", err)
	}
	defer rows.Close()
	var systems []domain.System
	for rows.Next() {
		var medium string
		var s domain.System
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Name, &medium, &s.Purpose, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan systems: %w", err)
		}
		s.Medium = domain.Medium(medium)
		systems = append(systems, s)
	}
	return systems, rows.Err()
}
