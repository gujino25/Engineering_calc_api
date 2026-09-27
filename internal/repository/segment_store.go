package repository

import (
	"cmp"
	"context"
	"enginer/internal/domain"
	"maps"
	"slices"
	"sync"
)

type SegmentStore struct {
	segments map[string]domain.Segment
	mtx      sync.RWMutex
}

func NewSegmentStore() *SegmentStore {

	return &SegmentStore{
		segments: make(map[string]domain.Segment),
	}
}

func (s *SegmentStore) Create(ctx context.Context, segment domain.Segment) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.segments[segment.ID]; ok {
		return domain.ErrSegmentAlreadyExists
	}
	s.segments[segment.ID] = segment

	return nil
}

func (s *SegmentStore) GetByID(ctx context.Context, id string) (domain.Segment, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	segments, ok := s.segments[id]
	if !ok {
		return domain.Segment{}, domain.ErrSegmentNotFound
	}
	return segments, nil
}

func (s *SegmentStore) ListBySystem(ctx context.Context, id string) []domain.Segment {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	tmp := make([]domain.Segment, 0, len(s.segments))

	for _, v := range s.segments {
		if v.SystemID == id {
			tmp = append(tmp, v)
		}
	}

	slices.SortFunc(tmp, func(a, b domain.Segment) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})

	return tmp
}

func (s *SegmentStore) List(ctx context.Context) map[string]domain.Segment {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	tmp := make(map[string]domain.Segment, len(s.segments))

	maps.Copy(tmp, s.segments)
	return tmp
}
