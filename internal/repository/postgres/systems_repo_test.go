//go:build integration

package postgres

import (
	"enginer/internal/domain"
	"errors"
	"testing"
	"time"
)

func createTestProject(t *testing.T) domain.Project {
	t.Helper()
	project := domain.NewProject("test", "testing")
	if err := NewProjectRepo(testPool).Create(t.Context(), project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	return project
}

func TestSystemRepo_CreateAndGetByID(t *testing.T) {
	t.Run("Валиидное создание и поиск", func(t *testing.T) {
		resetTables(t)
		repo := NewSystemRepo(testPool)
		project := createTestProject(t)
		system, err := domain.NewSystem(project.ID, "test", "air", "testing")
		if err != nil {
			t.Fatalf("system err = %v", err)
		}
		if err := repo.Create(t.Context(), system); err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := repo.GetByID(t.Context(), system.ID)

		if err != nil {
			t.Fatalf("got: %v", err)
		}

		if got.ID != system.ID {
			t.Errorf("ID = %v, ожидалось %v", got.ID, system.ID)
		}
		if got.Name != system.Name {
			t.Errorf("Name = %q, ожидалось %q", got.Name, system.Name)
		}
		if string(got.Medium) != string(system.Medium) {
			t.Errorf("Description = %q, ожидалось %q", got.Medium, system.Medium)
		}
		if got.Purpose != system.Purpose {
			t.Errorf("Purpose = %q, ожидалось %q", got.Purpose, system.Purpose)
		}
		if diff := got.CreatedAt.Sub(system.CreatedAt).Abs(); diff > time.Millisecond {
			t.Errorf("CreatedAt отличается на %v", diff)
		}
	})
	t.Run("Дубилкат", func(t *testing.T) {
		resetTables(t)
		repo := NewSystemRepo(testPool)
		project := createTestProject(t)
		system, err := domain.NewSystem(project.ID, "test", "air", "testing")
		if err != nil {
			t.Fatalf("system err = %v", err)
		}
		if err := repo.Create(t.Context(), system); err != nil {
			t.Fatalf("create: %v", err)
		}
		err = repo.Create(t.Context(), system)
		if !errors.Is(err, domain.ErrSystemAlreadyExists) {
			t.Fatalf("err =%v, ожидалось: %v", err, domain.ErrSystemAlreadyExists)
		}
	})
	t.Run("несуществующий проект", func(t *testing.T) {
		resetTables(t)
		repo := NewSystemRepo(testPool)
		system, err := domain.NewSystem("no_ID", "test", "air", "testing")
		if err != nil {
			t.Fatalf("system err = %v", err)
		}
		err = repo.Create(t.Context(), system)

		if !errors.Is(err, domain.ErrProjectNotFound) {
			t.Fatalf("create: %v", err)
		}
	})

	t.Run("Невалидный айди", func(t *testing.T) {
		resetTables(t)
		repo := NewSystemRepo(testPool)
		project := createTestProject(t)
		system, err := domain.NewSystem(project.ID, "test", "air", "testing")
		if err != nil {
			t.Fatalf("system err = %v", err)
		}
		if err := repo.Create(t.Context(), system); err != nil {
			t.Fatalf("create: %v", err)
		}
		_, err = repo.GetByID(t.Context(), "id_not_found")
		if !errors.Is(err, domain.ErrSystemNotFound) {
			t.Fatalf("err = %v, ожидалось %v", err, domain.ErrSystemNotFound)
		}
	})
}

func TestSystemRepo_List(t *testing.T) {
	resetTables(t)
	project := createTestProject(t)
	repo := NewSystemRepo(testPool)
	system1, err := domain.NewSystem(project.ID, "test1", "air", "testing1")
	if err != nil {
		t.Fatalf("system err = %v", err)
	}
	system2, err := domain.NewSystem(project.ID, "test2", "air", "testing2")
	if err != nil {
		t.Fatalf("system err = %v", err)
	}
	system2.CreatedAt = system1.CreatedAt.Add(-time.Millisecond)
	if err := repo.Create(t.Context(), system1); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.Create(t.Context(), system2); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repo.List(t.Context())
	if err != nil {
		t.Fatalf("got err :%v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %v, ожидалось 2", len(got))
	}

	if got[0].ID != system2.ID {
		t.Errorf("ID =%v, ожидалось: %v", got[0].ID, system2.ID)
	}
	if got[1].ID != system1.ID {
		t.Errorf("ID =%v, ожидалось: %v", got[1].ID, system1.ID)
	}
	if diff := got[0].CreatedAt.Sub(system2.CreatedAt).Abs(); diff > time.Millisecond {
		t.Errorf("CreatedAt отличается на %v", diff)
	}
}

func TestSystemRepo_GetByProject(t *testing.T) {
	t.Run("сущестувующий айди проекта", func(t *testing.T) {
		resetTables(t)
		project1 := createTestProject(t)
		project2 := createTestProject(t)
		repo := NewSystemRepo(testPool)
		system1, err := domain.NewSystem(project1.ID, "test1", "air", "valid projectID1")
		if err != nil {
			t.Fatalf("system creaton err = %v", err)
		}
		system2, err := domain.NewSystem(project2.ID, "test2", "air", "valid projectID2")
		if err != nil {
			t.Fatalf("system err = %v", err)
		}
		if err := repo.Create(t.Context(), system1); err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := repo.Create(t.Context(), system2); err != nil {
			t.Fatalf("create: %v", err)
		}

		got, err := repo.ListByProject(t.Context(), project1.ID)
		if len(got) != 1 {
			t.Fatalf("len(got) = %v, ожидалось 1", len(got))
		}
		if got[0].ProjectID != system1.ProjectID {
			t.Fatalf("projectID = %v, ожидалось %v", got[0].ProjectID, system1.ProjectID)
		}
	})
	t.Run("чужой проект", func(t *testing.T) {
		resetTables(t)
		project := createTestProject(t)
		repo := NewSystemRepo(testPool)
		system, err := domain.NewSystem(project.ID, "test", "air", "testing")

		if err != nil {
			t.Fatalf("system err = %v", err)
		}
		if err := repo.Create(t.Context(), system); err != nil {
			t.Fatalf("create: %v", err)
		}
		got, err := repo.ListByProject(t.Context(), "no_ID")
		if err != nil {
			t.Fatalf("err = %v, ожидалось: %v", err, nil)
		}
		if len(got) != 0 {
			t.Fatalf("len(got) = %v, ожидалось 0", len(got))
		}
	})
}
