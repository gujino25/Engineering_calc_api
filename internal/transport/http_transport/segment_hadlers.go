package httptransport

import (
	"encoding/json"
	"enginer/internal/domain"
	"enginer/internal/service"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

type SegmentHandlers struct {
	segmentService *service.SegmentService
}

func NewSegmentHandlers(store *service.SegmentService) *SegmentHandlers {
	return &SegmentHandlers{
		segmentService: store,
	}
}

func (s *SegmentHandlers) HandleCreateSegment(w http.ResponseWriter, r *http.Request) {
	var segmentDTO CreateSegmentDTO

	if err := json.NewDecoder(r.Body).Decode(&segmentDTO); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := segmentDTO.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	systemID := mux.Vars(r)["system_id"]
	var rect *domain.RectGeometry
	if segmentDTO.Rect != nil {
		rect = &domain.RectGeometry{
			Width:  segmentDTO.Rect.Width,
			Height: segmentDTO.Rect.Height,
		}
	}
	var round *domain.RoundGeometry
	if segmentDTO.Round != nil {
		round = &domain.RoundGeometry{
			Diameter: segmentDTO.Round.Diameter,
		}
	}
	ctx := r.Context()
	newSegment, err := s.segmentService.CreateSegment(
		ctx,
		systemID,
		segmentDTO.Name,
		domain.Shape(segmentDTO.Shape),
		rect,
		round,
		segmentDTO.Length)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSystemNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrSegmentAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, domain.ErrGeometryInvalid),
			errors.Is(err, domain.ErrLengthInvalid),
			errors.Is(err, domain.ErrShapeInvalid):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, toSegmentResponse(newSegment))

}

func (s *SegmentHandlers) HandleListBySystem(w http.ResponseWriter, r *http.Request) {
	systemID := mux.Vars(r)["system_id"]
	ctx := r.Context()
	segments, err := s.segmentService.ListBySystem(ctx, systemID)

	if err != nil {
		if errors.Is(err, domain.ErrSystemNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	res := make([]SegmentResponse, 0, len(segments))
	for _, v := range segments {
		res = append(res, toSegmentResponse(v))
	}
	writeJSON(w, http.StatusOK, res)
}
