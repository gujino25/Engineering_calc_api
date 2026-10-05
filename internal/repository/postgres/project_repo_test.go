//go:build integration

package postgres

import (
	"enginer/internal/domain"
	"errors"
	"testing"
	"time"
)

func TestProjectRepo_CreateAndGetByID(t *testing.T) {
	t.Run("валидное создание и поиск", func(t *testing.T) {
		resetTables(t)
		repo := NewProjectRepo(testPool)

		project := domain.NewProject("test", "air test")

		if err := repo.Create(t.Context(), project); err != nil {
			t.Fatalf("create: %v", err)
		}

		got, err := repo.GetByID(t.Context(), project.ID)

		if err != nil {
			t.Fatalf("get: %v", err)
		}

		if got.ID != project.ID {
			t.Errorf("ID = %v, ожидалось %v", got.ID, project.ID)
		}
		if got.Name != project.Name {
			t.Errorf("Name = %q, ожидалось %q", got.Name, project.Name)
		}
		if got.Description != project.Description {
			t.Errorf("Description = %q, ожидалось %q", got.Description, project.Description)
		}
		if diff := got.CreatedAt.Sub(project.CreatedAt).Abs(); diff > time.Millisecond {
			t.Errorf("CreatedAt отличается на %v", diff)
		}
	})
	t.Run("Дубликат при создании", func(t *testing.T) {
		resetTables(t)
		repo := NewProjectRepo(testPool)

		project := domain.NewProject("test", "testIDdublicate")
		if err := repo.Create(t.Context(), project); err != nil {
			t.Fatalf("create: %v", err)
		}
		err := repo.Create(t.Context(), project)
		if !errors.Is(err, domain.ErrProjectAlreadyExists) {
			t.Fatalf("err =%v, ожидалось: %v", err, domain.ErrProjectAlreadyExists)
		}
	})
	t.Run("несуществующий айдишник", func(t *testing.T) {
		resetTables(t)
		repo := NewProjectRepo(testPool)
		project := domain.NewProject("test", "testing")
		if err := repo.Create(t.Context(), project); err != nil {
			t.Fatalf("create: %v", err)
		}
		_, err := repo.GetByID(t.Context(), "несуществующий айди")

		if !errors.Is(err, domain.ErrProjectNotFound) {
			t.Fatalf("err = %v, ожидалось %v", err, domain.ErrProjectNotFound)
		}
	})
}

func TestProjectRepo_List(t *testing.T) {
	resetTables(t)
	repo := NewProjectRepo(testPool)

	project1 := domain.NewProject("test1", "testinglist1")
	project2 := domain.NewProject("test2", "testinglist2")
	project2.CreatedAt = project1.CreatedAt.Add(-time.Second)
	if err := repo.Create(t.Context(), project1); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := repo.Create(t.Context(), project2); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := repo.List(t.Context())

	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, ожидалось 2", len(got))
	}

	if got[0].ID != project2.ID {
		t.Errorf("ID =%v, ожидалось: %v", got[0].ID, project2.ID)
	}
	if got[1].ID != project1.ID {
		t.Errorf("ID =%v, ожидалось: %v", got[1].ID, project1.ID)
	}
	if diff := got[0].CreatedAt.Sub(project2.CreatedAt).Abs(); diff > time.Millisecond {
		t.Errorf("CreatedAt отличается на %v", diff)
	}
}
