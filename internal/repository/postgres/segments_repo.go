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

type SegmentRepo struct {
	pool *pgxpool.Pool
}

func NewSegmentRepo(pool *pgxpool.Pool) *SegmentRepo {
	return &SegmentRepo{
		pool: pool,
	}
}

func (s *SegmentRepo) Create(ctx context.Context, segment domain.Segment) error {
	const query = `INSERT INTO segments (id, system_id, name, shape, width, height, diameter, length, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`

	var width, height, diameter *int
	if segment.Rect != nil {
		width = &segment.Rect.Width
		height = &segment.Rect.Height
	}
	if segment.Round != nil {
		diameter = &segment.Round.Diameter
	}

	_, err := s.pool.Exec(ctx,
		query,
		segment.ID,
		segment.SystemID,
		segment.Name,
		string(segment.Shape),
		width,
		height,
		diameter,
		segment.Length,
		segment.CreatedAt)
	if err != nil {
		var pgrErr *pgconn.PgError
		if errors.As(err, &pgrErr) {
			switch pgrErr.Code {
			case "23505":
				return domain.ErrSegmentAlreadyExists
			case "23503":
				return domain.ErrSystemNotFound
			}
		}
		return fmt.Errorf("create segment: %w", err)
	}
	return nil
}

func (s *SegmentRepo) GetByID(ctx context.Context, id string) (domain.Segment, error) {
	const query = `SELECT id, system_id, name, shape, width, height, diameter, length, created_at FROM segments WHERE id = $1`

	segment, err := scanSegment(s.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Segment{}, domain.ErrSegmentNotFound
		}
		return domain.Segment{}, fmt.Errorf("get segment: %w", err)
	}
	return segment, nil
}

func (s *SegmentRepo) List(ctx context.Context) ([]domain.Segment, error) {
	const query = `SELECT id, system_id, name, shape, width, height, diameter, length, created_at FROM segments ORDER BY created_at, id`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all segments: %w", err)
	}
	defer rows.Close()

	var segments []domain.Segment
	for rows.Next() {
		segment, err := scanSegment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan segments: %w", err)
		}
		segments = append(segments, segment)
	}
	return segments, rows.Err()
}

func (s *SegmentRepo) ListBySystem(ctx context.Context, systemID string) ([]domain.Segment, error) {
	const query = `SELECT id, system_id, name, shape, width, height, diameter, length, created_at FROM segments WHERE system_id = $1 ORDER BY created_at, id`

	rows, err := s.pool.Query(ctx, query, systemID)
	if err != nil {
		return nil, fmt.Errorf("get all segments by system: %w", err)
	}
	defer rows.Close()

	var segments []domain.Segment
	for rows.Next() {
		segment, err := scanSegment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan segments: %w", err)
		}
		segments = append(segments, segment)
	}
	return segments, rows.Err()
}

func scanSegment(row pgx.Row) (domain.Segment, error) {
	var segment domain.Segment
	var shape string
	var width, height, diameter *int

	err := row.Scan(
		&segment.ID,
		&segment.SystemID,
		&segment.Name,
		&shape,
		&width,
		&height,
		&diameter,
		&segment.Length,
		&segment.CreatedAt,
	)
	if err != nil {
		return domain.Segment{}, err
	}

	segment.Shape = domain.Shape(shape)
	if width != nil && height != nil {
		segment.Rect = &domain.RectGeometry{Width: *width, Height: *height}
	}
	if diameter != nil {
		segment.Round = &domain.RoundGeometry{Diameter: *diameter}
	}
	return segment, nil
}
