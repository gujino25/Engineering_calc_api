//go:build integration

package postgres

import (
	"enginer/internal/domain"
	"errors"
	"testing"
	"time"
)

func createTestSystem(t *testing.T) domain.System {
	t.Helper()
	project := createTestProject(t)
	system, err := domain.NewSystem(project.ID, "test", "air", "testing")
	if err != nil {
		t.Fatalf("system err = %v", err)
	}
	if err := NewSystemRepo(testPool).Create(t.Context(), system); err != nil {
		t.Fatalf("create system: %v", err)
	}
	return system
}

func TestSegmentRepo_CreateAndGetByID(t *testing.T) {
	t.Run("валидный прямоугольный", func(t *testing.T) {
		resetTables(t)
		repo := NewSegmentRepo(testPool)
		system := createTestSystem(t)
		segment, err := domain.NewSegment(system.ID, "test", domain.ShapeRect, &domain.RectGeometry{Width: 400, Height: 200}, nil, 5.5, 1000)
		if err != nil {
			t.Fatalf("segment err = %v", err)
		}
		if err := repo.Create(t.Context(), segment); err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := repo.GetByID(t.Context(), segment.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.ID != segment.ID {
			t.Errorf("ID = %v, ожидалось %v", got.ID, segment.ID)
		}
		if got.Shape != domain.ShapeRect {
			t.Errorf("Shape = %v, ожидалось %v", got.Shape, domain.ShapeRect)
		}
		if got.Rect == nil || *got.Rect != *segment.Rect {
			t.Errorf("Rect = %+v, ожидалось %+v", got.Rect, segment.Rect)
		}
		if got.Round != nil {
			t.Errorf("Round = %+v, ожидалось nil", got.Round)
		}
		if got.Length != segment.Length {
			t.Errorf("Length = %v, ожидалось %v", got.Length, segment.Length)
		}
		if got.FlowRate != segment.FlowRate {
			t.Errorf("FlowRate = %v, ожидалось %v", got.FlowRate, segment.FlowRate)
		}
		if diff := got.CreatedAt.Sub(segment.CreatedAt).Abs(); diff > time.Millisecond {
			t.Errorf("CreatedAt отличается на %v", diff)
		}
	})
	t.Run("валидный круглый", func(t *testing.T) {
		resetTables(t)
		repo := NewSegmentRepo(testPool)
		system := createTestSystem(t)
		segment, err := domain.NewSegment(system.ID, "test", domain.ShapeRound, nil, &domain.RoundGeometry{Diameter: 315}, 3, 800)
		if err != nil {
			t.Fatalf("segment err = %v", err)
		}
		if err := repo.Create(t.Context(), segment); err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := repo.GetByID(t.Context(), segment.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Round == nil || got.Round.Diameter != 315 {
			t.Errorf("Round = %+v, ожидалось %+v", got.Round, segment.Round)
		}
		if got.Rect != nil {
			t.Errorf("Rect = %+v, ожидалось nil", got.Rect)
		}
	})
	t.Run("Дубликат", func(t *testing.T) {
		resetTables(t)
		repo := NewSegmentRepo(testPool)
		system := createTestSystem(t)
		segment, err := domain.NewSegment(system.ID, "test", domain.ShapeRound, nil, &domain.RoundGeometry{Diameter: 200}, 3, 800)
		if err != nil {
			t.Fatalf("segment err = %v", err)
		}
		if err := repo.Create(t.Context(), segment); err != nil {
			t.Fatalf("create: %v", err)
		}
		err = repo.Create(t.Context(), segment)
		if !errors.Is(err, domain.ErrSegmentAlreadyExists) {
			t.Fatalf("err = %v, ожидалось: %v", err, domain.ErrSegmentAlreadyExists)
		}
	})
	t.Run("несуществующая система", func(t *testing.T) {
		resetTables(t)
		repo := NewSegmentRepo(testPool)
		segment, err := domain.NewSegment("no_ID", "test", domain.ShapeRound, nil, &domain.RoundGeometry{Diameter: 200}, 3, 800)
		if err != nil {
			t.Fatalf("segment err = %v", err)
		}
		err = repo.Create(t.Context(), segment)
		if !errors.Is(err, domain.ErrSystemNotFound) {
			t.Fatalf("err = %v, ожидалось: %v", err, domain.ErrSystemNotFound)
		}
	})
	t.Run("Невалидный айди", func(t *testing.T) {
		resetTables(t)
		repo := NewSegmentRepo(testPool)
		_, err := repo.GetByID(t.Context(), "id_not_found")
		if !errors.Is(err, domain.ErrSegmentNotFound) {
			t.Fatalf("err = %v, ожидалось %v", err, domain.ErrSegmentNotFound)
		}
	})
}

func TestSegmentRepo_ListBySystem(t *testing.T) {
	resetTables(t)
	repo := NewSegmentRepo(testPool)
	system1 := createTestSystem(t)
	system2 := createTestSystem(t)

	segment1, err := domain.NewSegment(system1.ID, "test1", domain.ShapeRound, nil, &domain.RoundGeometry{Diameter: 200}, 3, 800)
	if err != nil {
		t.Fatalf("segment err = %v", err)
	}
	segment2, err := domain.NewSegment(system1.ID, "test2", domain.ShapeRound, nil, &domain.RoundGeometry{Diameter: 160}, 2, 500)
	if err != nil {
		t.Fatalf("segment err = %v", err)
	}
	segment2.CreatedAt = segment1.CreatedAt.Add(-time.Millisecond)
	otherSegment, err := domain.NewSegment(system2.ID, "other", domain.ShapeRound, nil, &domain.RoundGeometry{Diameter: 125}, 1, 300)
	if err != nil {
		t.Fatalf("segment err = %v", err)
	}
	for _, s := range []domain.Segment{segment1, segment2, otherSegment} {
		if err := repo.Create(t.Context(), s); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	got, err := repo.ListBySystem(t.Context(), system1.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %v, ожидалось 2", len(got))
	}
	if got[0].ID != segment2.ID {
		t.Errorf("ID = %v, ожидалось: %v", got[0].ID, segment2.ID)
	}
	if got[1].ID != segment1.ID {
		t.Errorf("ID = %v, ожидалось: %v", got[1].ID, segment1.ID)
	}

	all, err := repo.List(t.Context())
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("len(all) = %v, ожидалось 3", len(all))
	}
}

func TestSegmentRepo_FlowRateConstraint(t *testing.T) {
	resetTables(t)
	system := createTestSystem(t)
	_, err := testPool.Exec(t.Context(),
		`INSERT INTO segments (id, system_id, name, shape, diameter, length, flow_rate) VALUES ('bad', $1, 'bad', 'round', 200, 1, 0)`,
		system.ID)
	if err == nil {
		t.Fatalf("ожидалась ошибка CHECK constraint для flow_rate = 0")
	}
}
