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

	"github.com/alicebob/miniredis/v2"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func newTestSegmentHandlers(t *testing.T) (*SegmentHandlers, *repository.ProjectStore, *repository.SystemStore, *repository.SegmentStore) {
	t.Helper()

	mr := miniredis.RunT(t)
	cache := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { cache.Close() })
	projectStore := repository.NewProjectStore()
	systemStore := repository.NewSystemStore()
	segmentStore := repository.NewSegmentStore()
	segmentService := service.NewSegmentService(segmentStore, systemStore, cache)
	return NewSegmentHandlers(segmentService), projectStore, systemStore, segmentStore
}

func TestHandleCreateSegment(t *testing.T) {
	handlers, projectStore, systemStore, _ := newTestSegmentHandlers(t)

	project := domain.NewProject("тест", "проект для сегмента")
	if err := projectStore.Create(t.Context(), project); err != nil {
		t.Fatalf("не удалось создать проект: %v", err)
	}

	system, err := domain.NewSystem(project.ID, "система", "water", "перекачка воды")
	if err != nil {
		t.Fatalf("не удалось создать систему: %v", err)
	}
	if err := systemStore.Create(t.Context(), system); err != nil {
		t.Fatalf("не удалось положить систему в хранилище: %v", err)
	}

	tests := []struct {
		name       string
		systemID   string
		body       string
		wantStatus int
	}{
		{
			name:       "валидное создание прямоугольного воздуховода",
			systemID:   system.ID,
			body:       `{"name":"воздуховод 1","shape":"rect","rect":{"width":10,"height":20},"length":5}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "валидное создание круглой трубы",
			systemID:   system.ID,
			body:       `{"name":"труба 1","shape":"round","round":{"diameter":15},"length":5}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "пустое имя",
			systemID:   system.ID,
			body:       `{"name":"","shape":"round","round":{"diameter":15},"length":5}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "невалидная форма",
			systemID:   system.ID,
			body:       `{"name":"труба","shape":"треугольник","length":5}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "невалидная геометрия",
			systemID:   system.ID,
			body:       `{"name":"труба","shape":"rect","rect":{"width":0,"height":20},"length":5}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "невалидная длина",
			systemID:   system.ID,
			body:       `{"name":"труба","shape":"round","round":{"diameter":15},"length":0}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "битый json",
			systemID:   system.ID,
			body:       `не json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "несуществующая система",
			systemID:   "нет-такой-системы",
			body:       `{"name":"труба","shape":"round","round":{"diameter":15},"length":5}`,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/systems/"+tt.systemID+"/segments", strings.NewReader(tt.body))
			req = mux.SetURLVars(req, map[string]string{"system_id": tt.systemID})
			rec := httptest.NewRecorder()

			handlers.HandleCreateSegment(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("код = %v, ожидался %d, тело %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHandleListBySystem(t *testing.T) {
	t.Run("система не найдена", func(t *testing.T) {
		handlers, _, _, _ := newTestSegmentHandlers(t)
		req := httptest.NewRequest(http.MethodGet, "/systems/no_id/segments", nil)
		req = mux.SetURLVars(req, map[string]string{"system_id": "невалидный_айдишник"})
		rec := httptest.NewRecorder()

		handlers.HandleListBySystem(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("код = %v, ожидался %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("валидный список", func(t *testing.T) {
		handlers, projectStore, systemStore, segmentStore := newTestSegmentHandlers(t)

		project := domain.NewProject("тест", "проект для списка сегментов")
		if err := projectStore.Create(t.Context(), project); err != nil {
			t.Fatalf("не удалось создать проект: %v", err)
		}

		system, err := domain.NewSystem(project.ID, "система", "water", "перекачка воды")
		if err != nil {
			t.Fatalf("не удалось создать систему: %v", err)
		}
		if err := systemStore.Create(t.Context(), system); err != nil {
			t.Fatalf("не удалось положить систему в хранилище: %v", err)
		}

		segment, err := domain.NewSegment(system.ID, "воздуховод 1", domain.ShapeRect, &domain.RectGeometry{Width: 10, Height: 20}, nil, 5)
		if err != nil {
			t.Fatalf("не удалось создать сегмент: %v", err)
		}
		if err := segmentStore.Create(t.Context(), segment); err != nil {
			t.Fatalf("не удалось положить сегмент в хранилище: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/systems/"+system.ID+"/segments", nil)
		req = mux.SetURLVars(req, map[string]string{"system_id": system.ID})
		rec := httptest.NewRecorder()

		handlers.HandleListBySystem(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("код = %v, ожидался %d, тело %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var res []SegmentResponse
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("не удалось разобрать ответ: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("длина списка = %d, ожидалось 1", len(res))
		}
		if res[0].ID != segment.ID {
			t.Errorf("ID = %v, ожидался %v", res[0].ID, segment.ID)
		}
	})
}
