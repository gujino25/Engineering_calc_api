package httptransport

import (
	"context"
	"encoding/json"
	"enginer/internal/domain"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type projectStore interface {
	Create(ctx context.Context, project domain.Project) error
	GetByID(ctx context.Context, id string) (domain.Project, error)
	List(ctx context.Context) ([]domain.Project, error)
}

type ProjectHandlers struct {
	projectStore projectStore
}

func NewProjectHandlers(store projectStore) *ProjectHandlers {
	return &ProjectHandlers{
		projectStore: store,
	}
}

func (p *ProjectHandlers) HandleCreateProject(w http.ResponseWriter, r *http.Request) {
	var projectDTO CreateProjectDTO

	if err := json.NewDecoder(r.Body).Decode(&projectDTO); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return

	}

	if err := projectDTO.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	newProject := domain.NewProject(projectDTO.Name, projectDTO.Description)
	ctx := r.Context()
	if err := p.projectStore.Create(ctx, newProject); err != nil {
		if errors.Is(err, domain.ErrProjectAlreadyExists) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, toProjectResponse(newProject))
}

func (p *ProjectHandlers) HandleGetProject(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	ctx := r.Context()
	project, err := p.projectStore.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, toProjectResponse(project))
}

func (p *ProjectHandlers) HandleListProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projects, err := p.projectStore.List(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	res := make([]ProjectResponse, 0, len(projects))
	for _, v := range projects {
		res = append(res, toProjectResponse(v))

	}
	writeJSON(w, http.StatusOK, res)
}
