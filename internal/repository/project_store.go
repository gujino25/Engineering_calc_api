package repository

import (
	"cmp"
	"context"
	"enginer/internal/domain"
	"slices"
	"sync"
)

type ProjectStore struct {
	projects map[string]domain.Project
	mtx      sync.RWMutex
}

func NewProjectStore() *ProjectStore {
	return &ProjectStore{
		projects: make(map[string]domain.Project),
	}
}

func (p *ProjectStore) Create(ctx context.Context, project domain.Project) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if _, ok := p.projects[project.ID]; ok {
		return domain.ErrProjectAlreadyExists
	}
	p.projects[project.ID] = project

	return nil
}

func (p *ProjectStore) GetByID(ctx context.Context, id string) (domain.Project, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	project, ok := p.projects[id]

	if !ok {
		return domain.Project{}, domain.ErrProjectNotFound
	}
	return project, nil
}

func (p *ProjectStore) List(ctx context.Context) ([]domain.Project, error) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	tmp := make([]domain.Project, 0, len(p.projects))
	for _, v := range p.projects {
		tmp = append(tmp, v)
	}
	slices.SortFunc(tmp, func(a, b domain.Project) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})
	return tmp, nil
}
