package httptransport

import (
	"encoding/json"
	"enginer/internal/domain"
	"enginer/internal/repository"
	"enginer/internal/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func newTestSystemHandlers() (*SystemHandlers, *repository.ProjectStore, *repository.SystemStore) {
	systemStore := repository.NewSystemStore()
	projectStore := repository.NewProjectStore()
	systemService := service.NewSystemService(systemStore, projectStore)
	return NewSystemHandlers(systemService), projectStore, systemStore
}

func TestHandleCreateSystem(t *testing.T) {
	handlers, projectStore, _ := newTestSystemHandlers()

	project := domain.NewProject("тест", "проект для системы")
	if err := projectStore.Create(t.Context(), project); err != nil {
		t.Fatalf("не удалось создать проект: %v", err)
	}

	tests := []struct {
		name       string
		projectID  string
		body       string
		wantStatus int
	}{
		{
			name:       "валидное создание",
			projectID:  project.ID,
			body:       `{"name":"насос","medium":"water","purpose":"перекачка воды"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "пустое имя",
			projectID:  project.ID,
			body:       `{"name":"","medium":"water","purpose":"перекачка воды"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "невалидная среда",
			projectID:  project.ID,
			body:       `{"name":"насос","medium":"лава","purpose":"перекачка воды"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "битый json",
			projectID:  project.ID,
			body:       `не json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "несуществующий проект",
			projectID:  "нет-такого-проекта",
			body:       `{"name":"насос","medium":"water","purpose":"перекачка воды"}`,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/projects/"+tt.projectID+"/systems", strings.NewReader(tt.body))
			req = mux.SetURLVars(req, map[string]string{"project_id": tt.projectID})
			rec := httptest.NewRecorder()

			handlers.HandleCreateSystem(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("код = %v, ожидался %d, тело %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHandleListByProject(t *testing.T) {

	t.Run("проект не найден", func(t *testing.T) {
		handlers, _, _ := newTestSystemHandlers()
		req := httptest.NewRequest(http.MethodGet, "/projects/no_id/systems", nil)
		req = mux.SetURLVars(req, map[string]string{"project_id": "невалидный_айдишник"})
		rec := httptest.NewRecorder()
		handlers.HandleListByProject(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("код = %v, ожидался %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("валидный спискок", func(t *testing.T) {
		handlers, projectStore, systemStore := newTestSystemHandlers()
		project := domain.NewProject("валидный тест", "система с проектами")

		if err := projectStore.Create(t.Context(), project); err != nil {
			t.Fatalf("не удалось создать проект %v", err)
		}
		system, err := domain.NewSystem(project.ID, "система 1", "air", "просто так")
		if err != nil {
			t.Fatalf("не удалось создать систему %v", err)
		}
		if err := systemStore.Create(t.Context(), system); err != nil {
			t.Fatalf("не удалоь положить систему в хранилище %v", err)
		}
		req := httptest.NewRequest(http.MethodGet, "/projects/"+project.ID+"/systems", nil)
		req = mux.SetURLVars(req, map[string]string{"project_id": project.ID})
		rec := httptest.NewRecorder()
		handlers.HandleListByProject(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("код = %v, ожидался %d, тело %s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var res []SystemResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("не удалось разобрать ответ: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("длина списка = %d, ожидалось 1", len(res))
		}
		if res[0].ID != system.ID {
			t.Errorf("ID = %v, ожидался %v", res[0].ID, system.ID)
		}
	})
}
