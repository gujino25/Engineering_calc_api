package service

import (
	"context"
	"enginer/internal/domain"
	"enginer/internal/repository"
)

type projectGetter interface {
	GetByID(ctx context.Context, id string) (domain.Project, error)
}

type SystemService struct {
	systemStore  *repository.SystemStore
	projectStore projectGetter
}

func NewSystemService(systemStore *repository.SystemStore, projectStore projectGetter) *SystemService {
	return &SystemService{
		systemStore:  systemStore,
		projectStore: projectStore,
	}
}

func (s *SystemService) CreateSystem(ctx context.Context, projectID, name string, medium domain.Medium, purpose string) (domain.System, error) {
	if _, err := s.projectStore.GetByID(ctx, projectID); err != nil {
		return domain.System{}, err
	}

	system, err := domain.NewSystem(projectID, name, string(medium), purpose)
	if err != nil {
		return domain.System{}, err
	}

	if err := s.systemStore.Create(ctx, system); err != nil {
		return domain.System{}, err
	}

	return system, nil
}

func (s *SystemService) ListByProject(ctx context.Context, projectID string) ([]domain.System, error) {

	if _, err := s.projectStore.GetByID(ctx, projectID); err != nil {
		return nil, err
	}

	return s.systemStore.ListByProject(ctx, projectID), nil
}
