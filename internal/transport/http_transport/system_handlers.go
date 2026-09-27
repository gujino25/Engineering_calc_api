package httptransport

import (
	"encoding/json"
	"enginer/internal/domain"
	"enginer/internal/service"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type SystemHandlers struct {
	systemService *service.SystemService
}

func NewSystemHandlers(store *service.SystemService) *SystemHandlers {
	return &SystemHandlers{
		systemService: store,
	}
}

func (s *SystemHandlers) HandleCreateSystem(w http.ResponseWriter, r *http.Request) {
	var systemDTO CreateSystemDTO

	if err := json.NewDecoder(r.Body).Decode(&systemDTO); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := systemDTO.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	projectID := mux.Vars(r)["project_id"]
	ctx := r.Context()
	newSystem, err := s.systemService.CreateSystem(
		ctx,
		projectID,
		systemDTO.Name,
		domain.Medium(systemDTO.Medium),
		systemDTO.Purpose)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProjectNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrMediumInvalid):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrSystemAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toSystemResponse(newSystem))
}

func (s *SystemHandlers) HandleListByProject(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]
	ctx := r.Context()
	systems, err := s.systemService.ListByProject(ctx, projectID)

	if err != nil {
		if errors.Is(err, domain.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	res := make([]SystemResponse, 0, len(systems))
	for _, v := range systems {
		res = append(res, toSystemResponse(v))
	}
	writeJSON(w, http.StatusOK, res)
}
