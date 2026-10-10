package httptransport

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	projectHandlers *ProjectHandlers
	systemHandlers  *SystemHandlers
	segmentHandlers *SegmentHandlers
}

func NewHTTPServer(
	projectHandler *ProjectHandlers,
	systemHadlers *SystemHandlers,
	segmentHandlers *SegmentHandlers) *HTTPServer {
	return &HTTPServer{
		projectHandlers: projectHandler,
		systemHandlers:  systemHadlers,
		segmentHandlers: segmentHandlers,
	}
}

func (s *HTTPServer) StartServer(ctx context.Context, addr string) error {
	router := mux.NewRouter()
	router.Path("/projects").Methods("POST").HandlerFunc(s.projectHandlers.HandleCreateProject)
	router.Path("/projects/{id}").Methods("GET").HandlerFunc(s.projectHandlers.HandleGetProject)
	router.Path("/projects").Methods("GET").HandlerFunc(s.projectHandlers.HandleListProjects)
	router.Path("/projects/{project_id}/systems").Methods("POST").HandlerFunc(s.systemHandlers.HandleCreateSystem)
	router.Path("/projects/{project_id}/systems").Methods("GET").HandlerFunc(s.systemHandlers.HandleListByProject)
	router.Path("/systems/{system_id}/segments").Methods("POST").HandlerFunc(s.segmentHandlers.HandleCreateSegment)
	router.Path("/systems/{system_id}/segments").Methods("GET").HandlerFunc(s.segmentHandlers.HandleListBySystem)

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
	return nil
}
