package repository

import (
	"enginer/internal/domain"
	"errors"
	"testing"
)

func TestProjectCreateAndGetByID(t *testing.T) {
	store := NewProjectStore()

	project := domain.NewProject("Дача", "Вентиляция на даче")

	err := store.Create(t.Context(), project)

	if err != nil {
		t.Fatalf("Не удалось создать в сторе %v", err)
	}

	got, err := store.GetByID(t.Context(), project.ID)

	if err != nil {
		t.Fatalf("Не удалось получить систему %v", err)
	}

	if got.Name != project.Name {
		t.Fatalf("Name =%q, ожидалось %q", got.Name, project.Name)
	}
	if got.Description != project.Description {
		t.Fatalf("Description =%q, ожидалось %q", got.Description, project.Description)
	}

}

func TestGotByID_NotFound(t *testing.T) {
	store := NewProjectStore()

	_, err := store.GetByID(t.Context(), "adadad")

	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Fatalf("err = %v, ожидалась ErrProjectNotFound:", err)
	}
}

func TestList(t *testing.T) {
	store := NewProjectStore()

	project1 := domain.NewProject("Дача", "Вентиляция на даче")
	project2 := domain.NewProject("Дача", "Водоснабжение на даче")

	store.Create(t.Context(), project1)
	store.Create(t.Context(), project2)

	got := store.List(t.Context())

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, ожидалось 2", len(got))
	}

	check := make(map[string]bool, len(got))
	for _, p := range got {
		check[p.ID] = true
	}

	if !check[project1.ID] {
		t.Errorf("проект %s не найден в рузльтате List", project1.ID)
	}

	if !check[project2.ID] {
		t.Errorf("проект %s не найден в рузльтате List", project2.ID)
	}

}
