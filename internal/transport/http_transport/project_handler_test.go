package httptransport

import (
	"encoding/json"
	"enginer/internal/domain"
	"enginer/internal/repository"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func newTestProjectHandlers() *ProjectHandlers {
	store := repository.NewProjectStore()
	return NewProjectHandlers(store)
}

func TestHandleCreateProject(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "валидное создание",
			body:       `{"name":"Дача","description":"тест"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "пустое имя",
			body:       `{"name":"","description":"тест"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "битый json",
			body:       `не json`,
			wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlers := newTestProjectHandlers()
			req := httptest.NewRequest(http.MethodPost, "/project", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			handlers.HandleCreateProject(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("код = %v, ожидался %d, тело %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHandleGetProject(t *testing.T) {
	store := repository.NewProjectStore()
	handlers := NewProjectHandlers(store)
	project := domain.NewProject("тест", "тест по айди")
	store.Create(t.Context(), project)
	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "верный айди",
			id:         project.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "невалидный айди",
			id:         "нет-такого-айди",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/projects/"+tt.id, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.id})
			rec := httptest.NewRecorder()

			handlers.HandleGetProject(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("код = %v, ожидался %d, тело %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}

}

func TestHandleListProjects(t *testing.T) {
	t.Run("пустой список", func(t *testing.T) {
		handlers := newTestProjectHandlers()

		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		rec := httptest.NewRecorder()

		handlers.HandleListProjects(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("код = %v, ожидался %d, тело %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var res []ProjectResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("не удалось распарсить ответ: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("длина списка = %d, ожидалось 0", len(res))
		}
	})

	t.Run("несколько проектов", func(t *testing.T) {
		store := repository.NewProjectStore()
		handlers := NewProjectHandlers(store)

		project1 := domain.NewProject("test", "project1 test")
		project2 := domain.NewProject("project2", "project2 test")
		if err := store.Create(t.Context(), project1); err != nil {
			t.Fatalf("не удалось создать проект1: %v", err)
		}
		if err := store.Create(t.Context(), project2); err != nil {
			t.Fatalf("не удалось создать проект2: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		rec := httptest.NewRecorder()

		handlers.HandleListProjects(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("код = %v, ожидался %d, тело %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var res []ProjectResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("не удалось распарсить ответ: %v", err)
		}
		if len(res) != 2 {
			t.Fatalf("длина списка = %d, ожидалось 2", len(res))
		}

		gotIDs := map[string]bool{res[0].ID: true, res[1].ID: true}
		if !gotIDs[project1.ID] || !gotIDs[project2.ID] {
			t.Errorf("в ответе нет ожидаемых айди, получено: %+v", res)
		}
	})
}
