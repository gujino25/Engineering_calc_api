package service

import (
	"enginer/internal/domain"
	"enginer/internal/repository"
	"errors"
	"testing"
)

func newTestSystemService() (*SystemService, *repository.ProjectStore) {
	projectStore := repository.NewProjectStore()
	svc := &SystemService{
		systemStore:  repository.NewSystemStore(),
		projectStore: projectStore,
	}
	return svc, projectStore
}

func TestSystemService_CreateSystem(t *testing.T) {
	t.Run("проект не найден", func(t *testing.T) {
		svc, _ := newTestSystemService()

		_, err := svc.CreateSystem(t.Context(), "no-such-project", "Система 1", domain.MediumAir, "офис")
		if !errors.Is(err, domain.ErrProjectNotFound) {
			t.Fatalf("err = %v, ожидалось %v", err, domain.ErrProjectNotFound)
		}
	})

	t.Run("невалидная среда", func(t *testing.T) {
		svc, projectStore := newTestSystemService()

		project := domain.NewProject("Проект 1", "описание")
		if err := projectStore.Create(t.Context(), project); err != nil {
			t.Fatalf("не удалось создать проект: %v", err)
		}

		_, err := svc.CreateSystem(t.Context(), project.ID, "Система 1", domain.Medium("газ"), "офис")
		if !errors.Is(err, domain.ErrMediumInvalid) {
			t.Fatalf("err = %v, ожидалось %v", err, domain.ErrMediumInvalid)
		}
	})

	t.Run("успешное создание", func(t *testing.T) {
		svc, projectStore := newTestSystemService()

		project := domain.NewProject("Проект 1", "описание")
		if err := projectStore.Create(t.Context(), project); err != nil {
			t.Fatalf("не удалось создать проект: %v", err)
		}

		system, err := svc.CreateSystem(t.Context(), project.ID, "Система 1", domain.MediumAir, "офис")
		if err != nil {
			t.Fatalf("не удалось создать систему: %v", err)
		}
		if system.ProjectID != project.ID {
			t.Errorf("ProjectID = %v, ожидалось %v", system.ProjectID, project.ID)
		}
		if system.Name != "Система 1" {
			t.Errorf("Name = %v, ожидалось %v", system.Name, "Система 1")
		}

		got, err := svc.systemStore.GetByID(t.Context(), system.ID)
		if err != nil {
			t.Fatalf("система не найдена в хранилище: %v", err)
		}
		if got.ID != system.ID {
			t.Errorf("ID = %v, ожидалось %v", got.ID, system.ID)
		}
	})
}

func TestSystemService_ListByProject(t *testing.T) {
	t.Run("проект не найден", func(t *testing.T) {
		svc, _ := newTestSystemService()

		_, err := svc.ListByProject(t.Context(), "no-such-project")
		if !errors.Is(err, domain.ErrProjectNotFound) {
			t.Fatalf("err = %v, ожидалось %v", err, domain.ErrProjectNotFound)
		}
	})

	t.Run("возвращает только системы своего проекта", func(t *testing.T) {
		svc, projectStore := newTestSystemService()

		project1 := domain.NewProject("Проект 1", "описание")
		project2 := domain.NewProject("Проект 2", "описание")
		if err := projectStore.Create(t.Context(), project1); err != nil {
			t.Fatalf("не удалось создать проект 1: %v", err)
		}
		if err := projectStore.Create(t.Context(), project2); err != nil {
			t.Fatalf("не удалось создать проект 2: %v", err)
		}

		system1, err := svc.CreateSystem(t.Context(), project1.ID, "Система 1", domain.MediumAir, "офис")
		if err != nil {
			t.Fatalf("не удалось создать систему 1: %v", err)
		}
		system2, err := svc.CreateSystem(t.Context(), project1.ID, "Система 2", domain.MediumWater, "склад")
		if err != nil {
			t.Fatalf("не удалось создать систему 2: %v", err)
		}
		otherSystem, err := svc.CreateSystem(t.Context(), project2.ID, "Чужая система", domain.MediumAir, "офис")
		if err != nil {
			t.Fatalf("не удалось создать чужую систему: %v", err)
		}

		got, err := svc.ListByProject(t.Context(), project1.ID)
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
		if !check[system1.ID] {
			t.Errorf("система %s не найдена в результате", system1.ID)
		}
		if !check[system2.ID] {
			t.Errorf("система %s не найдена в результате", system2.ID)
		}
		if check[otherSystem.ProjectID] {
			t.Errorf("система %s из другого проекта не должна быть в результате", otherSystem.ID)
		}
	})
}
