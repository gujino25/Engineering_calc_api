package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Shape string

const (
	ShapeRect  = "rect"
	ShapeRound = "round"
)

type RectGeometry struct {
	Width  int
	Height int
}

type RoundGeometry struct {
	Diameter int
}

type Segment struct {
	ID        string
	SystemID  string
	Name      string
	Shape     Shape
	Rect      *RectGeometry
	Round     *RoundGeometry
	Length    float64
	FlowRate  float64
	CreatedAt time.Time
}

func (s Shape) IsValid() bool {
	switch s {
	case ShapeRect, ShapeRound:
		return true
	}
	return false
}

func NewSegment(systemID string, name string, shape Shape, rect *RectGeometry, round *RoundGeometry, length float64, flowRate float64) (Segment, error) {
	s := Shape(shape)

	if !s.IsValid() {
		return Segment{}, ErrShapeInvalid
	}
	switch shape {
	case ShapeRect:
		if round != nil {
			return Segment{}, ErrGeometryInvalid
		}
		if rect == nil {
			return Segment{}, ErrGeometryInvalid
		}
		if rect.Height <= 0 || rect.Width <= 0 {
			return Segment{}, ErrGeometryInvalid
		}
	case ShapeRound:
		if rect != nil {
			return Segment{}, ErrGeometryInvalid
		}
		if round == nil {
			return Segment{}, ErrGeometryInvalid
		}
		if round.Diameter <= 0 {
			return Segment{}, ErrGeometryInvalid
		}
	}
	if flowRate <= 0 {
		return Segment{}, ErrFlowRateInvalid
	}
	if length <= 0 {
		return Segment{}, ErrLengthInvalid
	}

	return Segment{
		ID:        uuid.NewString(),
		SystemID:  systemID,
		Name:      name,
		Shape:     shape,
		Rect:      rect,
		Round:     round,
		Length:    length,
		FlowRate:  flowRate,
		CreatedAt: time.Now(),
	}, nil
}

func (s Segment) Area() (float64, error) {
	switch s.Shape {
	case ShapeRect:
		if s.Rect == nil {
			return 0, ErrGeometryInvalid
		}
		w := float64(s.Rect.Width) / 1000
		h := float64(s.Rect.Height) / 1000
		return w * h, nil
	case ShapeRound:
		if s.Round == nil {
			return 0, ErrGeometryInvalid
		}
		d := float64(s.Round.Diameter) / 1000
		return math.Pi * d * d / 4, nil
	default:
		return 0, ErrUnknownShape
	}
}

func (s Segment) Velocity() (float64, error) {
	f, err := s.Area()
	if err != nil {
		return 0, err
	}
	if f <= 0 {
		return 0, ErrAreaInvalid
	}
	return s.FlowRate / (3600 * f), nil
}
