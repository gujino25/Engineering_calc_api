package domain

import (
	"errors"
	"math"
	"testing"
)

func TestNewSegment(t *testing.T) {
	tests := []struct {
		name     string
		shape    Shape
		rect     *RectGeometry
		round    *RoundGeometry
		flowRate float64
		length   float64
		wantErr  error
	}{
		{
			name:     "валидный прямоугольный",
			shape:    ShapeRect,
			rect:     &RectGeometry{Width: 400, Height: 250},
			round:    nil,
			flowRate: 1.5,
			length:   5.2,
			wantErr:  nil,
		},
		{
			name:     "валидный круглый",
			shape:    ShapeRound,
			rect:     nil,
			round:    &RoundGeometry{Diameter: 200},
			flowRate: 2.5,
			length:   3.1,
			wantErr:  nil,
		},
		{
			name:     "прямоугольный без геометрии",
			shape:    ShapeRect,
			rect:     nil,
			round:    nil,
			length:   5.2,
			flowRate: 3.5,
			wantErr:  ErrGeometryInvalid,
		},
		{
			name:     "противоречие: rect с round одновременно",
			shape:    ShapeRect,
			rect:     &RectGeometry{Width: 400, Height: 250},
			round:    &RoundGeometry{Diameter: 200},
			flowRate: 5,
			length:   5.2,
			wantErr:  ErrGeometryInvalid,
		},
		{
			name:     "нулевая длина",
			shape:    ShapeRect,
			rect:     &RectGeometry{Width: 400, Height: 250},
			round:    nil,
			flowRate: 2.5,
			length:   0,
			wantErr:  ErrLengthInvalid,
		},
		{
			name:     "невалидная форма",
			shape:    "invalid",
			rect:     nil,
			round:    nil,
			flowRate: 2.5,
			length:   5.2,
			wantErr:  ErrShapeInvalid,
		},
		{
			name:  "невалидный расход",
			shape: ShapeRect,
			rect: &RectGeometry{
				Width:  100,
				Height: 100,
			},
			round:    nil,
			flowRate: 0,
			length:   5.2,
			wantErr:  ErrFlowRateInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSegment("system-1", "тест", tt.shape, tt.rect, tt.round, tt.length, tt.flowRate)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, ожидалось %v", err, tt.wantErr)
			}
		})
	}
}

func TestVelosityAndArea(t *testing.T) {
	test := []struct {
		name    string
		seg     Segment
		want    float64
		wantErr error
	}{
		{
			name: "rect 400x200 at 1000 м3/ч",
			seg:  Segment{Shape: ShapeRect, Rect: &RectGeometry{Width: 400, Height: 200}, FlowRate: 1000},
			want: 3.47,
		},
		{
			name: "round 315 at 1000 м3/ч",
			seg:  Segment{Shape: ShapeRound, Round: &RoundGeometry{Diameter: 315}, FlowRate: 1000},
			want: 3.56,
		},
		{
			name:    "невалидная геометрия",
			seg:     Segment{Shape: ShapeRect, FlowRate: 1000},
			wantErr: ErrGeometryInvalid,
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.seg.Velocity()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want = %v, get = %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && math.Abs(got-tt.want) > 0.01 {
				t.Errorf("got %.4f, want %.4f", got, tt.want)
			}
		})
	}
}
