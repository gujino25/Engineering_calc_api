package service

import (
	"enginer/internal/domain"
	"enginer/internal/repository"
	"errors"
	"testing"
)

func newTestSegmentService() (*SegmentService, *repository.SystemStore) {
	systemstore := repository.NewSystemStore()
	svc := &SegmentService{
		segmentStore: repository.NewSegmentStore(),
		systemStore:  systemstore,
	}
	return svc, systemstore
}

func newTestSegmentServiceWithSystem(t *testing.T) (*SegmentService, domain.System) {
	t.Helper()

	svc, systemStore := newTestSegmentService()
	system, err := domain.NewSystem("project-1", "Вытяжка", "air", "Вытяжка на дачу")
	if err != nil {
		t.Fatalf("не удалось создать систему: %v", err)
	}
	if err := systemStore.Create(t.Context(), system); err != nil {
		t.Fatalf("не удалось создать хранилище систем: %v", err)
	}
	return svc, system
}

func TestSegmentService_CreateSegment(t *testing.T) {
	t.Run("система не найдена", func(t *testing.T) {
		svc, _ := newTestSegmentService()
		_, err := svc.CreateSegment(t.Context(), "system-1", "От вру", "rect", &domain.RectGeometry{Width: 200, Height: 150}, nil, 5.8)
		if !errors.Is(err, domain.ErrSystemNotFound) {
			t.Fatalf("err = %v, ожидалось %v", err, domain.ErrSystemNotFound)
		}

	})
	t.Run("Форма и геометрия не совпали", func(t *testing.T) {
		svc, system := newTestSegmentServiceWithSystem(t)
		_, err := svc.CreateSegment(t.Context(), system.ID, "От вру", "round", &domain.RectGeometry{Width: 200, Height: 150}, nil, 5.8)
		if !errors.Is(err, domain.ErrGeometryInvalid) {
			t.Fatalf("err = %v, ожидалось %v", err, domain.ErrGeometryInvalid)
		}
	})
	t.Run("Невалидная геометрия", func(t *testing.T) {
		svc, system := newTestSegmentServiceWithSystem(t)
		_, err := svc.CreateSegment(t.Context(), system.ID, "От вру", "rect", &domain.RectGeometry{Width: 0, Height: 150}, nil, 5.8)
		if !errors.Is(err, domain.ErrGeometryInvalid) {
			t.Fatalf("err = %v, ожидалось %v", err, domain.ErrGeometryInvalid)
		}
	})

	t.Run("Невалидная форма", func(t *testing.T) {
		svc, system := newTestSegmentServiceWithSystem(t)
		_, err := svc.CreateSegment(t.Context(), system.ID, "От вру", "жопа", &domain.RectGeometry{Width: 250, Height: 150}, nil, 5.8)
		if !errors.Is(err, domain.ErrShapeInvalid) {
			t.Fatalf("err = %v, ожидалось %v", err, domain.ErrShapeInvalid)
		}

	})
	t.Run("Нулевая длина", func(t *testing.T) {
		svc, system := newTestSegmentServiceWithSystem(t)
		_, err := svc.CreateSegment(t.Context(), system.ID, "От вру", "rect", &domain.RectGeometry{Width: 200, Height: 150}, nil, 0)
		if !errors.Is(err, domain.ErrLengthInvalid) {
			t.Fatalf(" err = %v, ожидалось %v", err, domain.ErrLengthInvalid)
		}
	})
	t.Run("валидное создание сегмента", func(t *testing.T) {
		svc, system := newTestSegmentServiceWithSystem(t)
		segment, err := svc.CreateSegment(t.Context(), system.ID, "После вру 1", "round", nil, &domain.RoundGeometry{Diameter: 250}, 5.8)
		if err != nil {
			t.Fatalf("Не удалось создать сегмент %v", err)
		}
		if segment.SystemID != system.ID {
			t.Fatalf("SystemID = %v, ожидалось = %v", segment.SystemID, system.ID)
		}
		if segment.Name != "После вру 1" {
			t.Fatalf("Name = %v, ожидалось = %v", segment.Name, "После вру 1")
		}

		got, err := svc.segmentStore.(*repository.SegmentStore).GetByID(t.Context(), segment.ID)
		if err != nil {
			t.Fatalf("Сегмент не найден в хранилище %v", err)
		}
		if got.ID != segment.ID {
			t.Fatalf("ID = %v, ожидалось %v", got.ID, segment.ID)
		}
	})
}

func TestSegmentService_ListBySystem(t *testing.T) {
	t.Run("Нет систем", func(t *testing.T) {
		ctx := t.Context()
		svc, _ := newTestSegmentService()
		_, err := svc.ListBySystem(ctx, "no system at all")
		if !errors.Is(err, domain.ErrSystemNotFound) {
			t.Fatalf("Err = %v, ожидалось %v", err, domain.ErrSystemNotFound)
		}

	})
	t.Run("Валидный тест по системе", func(t *testing.T) {
		ctx := t.Context()
		svc, systemStore := newTestSegmentService()
		system1, err := domain.NewSystem("project-1", "Вытяжка", "air", "Вытяжка с кухни")
		if err != nil {
			t.Fatalf("Не удалось создать систему %v", err)
		}
		system2, err := domain.NewSystem("project-2", "Приток", "air", "Приток в общий зал")
		if err != nil {
			t.Fatalf("Не удалось создать систему %v", err)
		}

		if err := systemStore.Create(t.Context(), system1); err != nil {
			t.Fatalf("Не удалось создать хранилище систем  %v", err)
		}
		if err := systemStore.Create(t.Context(), system2); err != nil {
			t.Fatalf("Не удалось создать хранилище систем  %v", err)
		}

		segment1, err := svc.CreateSegment(ctx, system1.ID, "От вру 1", "round", nil, &domain.RoundGeometry{Diameter: 250}, 1.2)
		if err != nil {
			t.Fatalf("Не удалось создать сегмент %v", err)
		}
		segment2, err := svc.CreateSegment(ctx, system1.ID, "От вру 1", "round", nil, &domain.RoundGeometry{Diameter: 160}, 1.5)
		if err != nil {
			t.Fatalf("Не удалось создать сегмент %v", err)
		}
		otherSegment, err := svc.CreateSegment(ctx, system2.ID, "От вру 2", "round", nil, &domain.RoundGeometry{Diameter: 125}, 1.7)
		if err != nil {
			t.Fatalf("Не удалось создать сегмент %v", err)
		}
		got, err := svc.ListBySystem(ctx, system1.ID)
		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(got) = %d, ожидалось 2", len(got))
		}
		check := make(map[string]bool, len(got))
		for _, p := range got {
			check[p.ID] = true
		}
		if !check[segment1.ID] {
			t.Errorf("сегмент %s не найдена в результате", segment1.ID)
		}
		if !check[segment2.ID] {
			t.Errorf("сегмент %s не найдена в результате", segment2.ID)
		}
		if check[otherSegment.ID] {
			t.Errorf("сегмент %s из другой системы не должен быть в результате", otherSegment.ID)
		}
	})
}
