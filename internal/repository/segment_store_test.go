package repository

import (
	"enginer/internal/domain"
	"errors"
	"testing"
)

func TestSegmentStore_CreateAndGetByID(t *testing.T) {
	tests := []struct {
		id     string
		name   string
		shape  domain.Shape
		rect   *domain.RectGeometry
		round  *domain.RoundGeometry
		length float64
	}{
		{
			id:     "segment1",
			name:   "прямоугольный",
			shape:  domain.ShapeRect,
			rect:   &domain.RectGeometry{Width: 300, Height: 200},
			round:  nil,
			length: 2.5,
		},
		{
			id:     "segment2",
			name:   "круглый",
			shape:  domain.ShapeRound,
			rect:   nil,
			round:  &domain.RoundGeometry{Diameter: 250},
			length: 1.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			store := NewSegmentStore()
			segment, err := domain.NewSegment(tt.id, tt.name, tt.shape, tt.rect, tt.round, tt.length)
			if err != nil {
				t.Fatalf("не удалось создать сегмент %v", err)
			}
			if err := store.Create(t.Context(), segment); err != nil {
				t.Fatalf("не удалось создать харанилище %v", err)
			}

			got, err := store.GetByID(t.Context(), segment.ID)
			if err != nil {
				t.Fatalf("не удалось найти сегмент: %v", err)
			}
			if got.Shape != segment.Shape {
				t.Errorf("Shape = %v, ожидалось %v", got.Shape, segment.Shape)
			}
			if tt.shape == domain.ShapeRect && got.Rect.Width != tt.rect.Width {
				t.Errorf("Width = %v, ожидалась %v", got.Rect.Width, tt.rect.Width)
			}
			if tt.shape == domain.ShapeRound && got.Round.Diameter != tt.round.Diameter {
				t.Errorf("Diameter = %v, ожидалась %v", got.Round.Diameter, tt.round.Diameter)
			}
		})
	}

}

func TestSegmentStore_GetByID_NotFound(t *testing.T) {
	segments := NewSegmentStore()

	_, err := segments.GetByID(t.Context(), "aaaa")

	if !errors.Is(err, domain.ErrSegmentNotFound) {
		t.Fatalf("err = %v, ожидалась ErrSegmentNotFound", err)
	}
}

func TestSegmentStore_List(t *testing.T) {
	store := NewSegmentStore()

	segment1, _ := domain.NewSegment("system-1", "После вру 1", "rect", &domain.RectGeometry{Width: 200, Height: 300}, nil, 3.5)
	segment2, _ := domain.NewSegment("syetem-2", "После вру 2", "rect", &domain.RectGeometry{Width: 100, Height: 100}, nil, 2.5)
	store.Create(t.Context(), segment1)
	store.Create(t.Context(), segment2)

	got := store.List(t.Context())
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, ожидалось 2", len(got))
	}

	if _, ok := got[segment1.ID]; !ok {
		t.Errorf("сегмент %s не найден в рузльтате List", segment1.ID)
	}

	if _, ok := got[segment2.ID]; !ok {
		t.Errorf("сегмент %s не найден в рузльтате List", segment2.ID)
	}
}

func TestSegment_ListBySystem(t *testing.T) {
	store := NewSegmentStore()

	segment1, err := domain.NewSegment("system-1", "После вру 1", "rect", &domain.RectGeometry{Width: 200, Height: 300}, nil, 3.5)
	if err != nil {
		t.Fatalf("Не удалось создать сегмент1 %v", err)
	}
	segment2, err := domain.NewSegment("system-2", "После вру 1", "rect", &domain.RectGeometry{Width: 300, Height: 200}, nil, 0.5)
	if err != nil {
		t.Fatalf("Не удалось создать сегмент2 %v", err)
	}
	segment3, err := domain.NewSegment("system-2", "После вру 2", "rect", &domain.RectGeometry{Width: 200, Height: 150}, nil, 1.5)
	if err != nil {
		t.Fatalf("Не удалось создать сегмент3 %v", err)
	}

	store.Create(t.Context(), segment1)
	store.Create(t.Context(), segment2)
	store.Create(t.Context(), segment3)

	got, err := store.ListBySystem(t.Context(), "system-2")
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, ожидалось 2", len(got))
	}

}
