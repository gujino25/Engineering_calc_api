package repository

import (
	"cmp"
	"context"
	"enginer/internal/domain"
	"maps"
	"slices"
	"sync"
)

type SystemStore struct {
	systems map[string]domain.System
	mtx     sync.RWMutex
}

func NewSystemStore() *SystemStore {

	return &SystemStore{
		systems: make(map[string]domain.System),
	}

}

func (s *SystemStore) Create(ctx context.Context, system domain.System) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.systems[system.ID]; ok {
		return domain.ErrSystemAlreadyExists
	}
	s.systems[system.ID] = system

	return nil
}

func (s *SystemStore) GetByID(ctx context.Context, id string) (domain.System, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	system, ok := s.systems[id]

	if !ok {
		return domain.System{}, domain.ErrSystemNotFound
	}

	return system, nil
}

func (s *SystemStore) ListByProject(ctx context.Context, id string) ([]domain.System, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	tmp := make([]domain.System, 0, len(s.systems))

	for _, v := range s.systems {
		if v.ProjectID == id {
			tmp = append(tmp, v)
		}
	}

	slices.SortFunc(tmp, func(a, b domain.System) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})

	return tmp, nil
}

func (s *SystemStore) List(ctx context.Context) map[string]domain.System {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	tmp := make(map[string]domain.System, len(s.systems))

	maps.Copy(tmp, s.systems)

	return tmp
}
