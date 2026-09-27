package service

import (
	"context"
	"enginer/internal/domain"
)

type systemGetter interface {
	GetByID(ctx context.Context, id string) (domain.System, error)
}

type segmentStore interface {
	Create(ctx context.Context, segment domain.Segment) error
	ListBySystem(ctx context.Context, systemID string) ([]domain.Segment, error)
}

type SegmentService struct {
	segmentStore segmentStore
	systemStore  systemGetter
}

func NewSegmentService(segmentStore segmentStore, systemStore systemGetter) *SegmentService {
	return &SegmentService{
		segmentStore: segmentStore,
		systemStore:  systemStore,
	}
}

func (s *SegmentService) CreateSegment(ctx context.Context, systemID, name string, shape domain.Shape, rect *domain.RectGeometry, round *domain.RoundGeometry, length float64) (domain.Segment, error) {
	if _, err := s.systemStore.GetByID(ctx, systemID); err != nil {
		return domain.Segment{}, err
	}
	segment, err := domain.NewSegment(systemID, name, shape, rect, round, length)
	if err != nil {
		return domain.Segment{}, err
	}

	if err := s.segmentStore.Create(ctx, segment); err != nil {
		return domain.Segment{}, err
	}

	return segment, nil
}

func (s *SegmentService) ListBySystem(ctx context.Context, systemID string) ([]domain.Segment, error) {
	if _, err := s.systemStore.GetByID(ctx, systemID); err != nil {
		return nil, err
	}
	return s.segmentStore.ListBySystem(ctx, systemID)
}
