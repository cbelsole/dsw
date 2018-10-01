package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/cbelsole/dsw/db"
	"github.com/cbelsole/dsw/processors"
	"github.com/cbelsole/dsw/types"
)

type (
	Handler struct {
		DB  *db.DB
		Job processors.Job
	}
	createJobRequest struct {
		ErrorURI  *string                `json:"error_uri"`
		ExecuteAt time.Time              `json:"execute_at"`
		Payload   map[string]interface{} `json:"payload"`
		URI       string                 `json:"uri"`
	}
)

// HealthHandler returns a 200 if the service is healthy and a 500 if it is not
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.Ping(); err != nil {
		log.Printf("health check failed: %s\n", err)
		writeHTTPResponse(w, http.StatusInternalServerError, map[string]string{"message": "I'm unhealthy"})
	} else {
		writeHTTPResponse(w, http.StatusOK, map[string]string{"message": "I'm healthy"})
	}
}

// CreateJob takes a createJobRequest and enqueues the job for processing.
func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req createJobRequest
	if err := decoder.Decode(&req); err != nil {
		writeHTTPError(w, http.StatusBadRequest, err)
		return
	}

	// validate URI
	if _, err := url.ParseRequestURI(req.URI); err != nil {
		writeHTTPError(w, http.StatusBadRequest, err)
		return
	}

	// validate error URI if present
	if req.ErrorURI != nil {
		if _, err := url.ParseRequestURI(*req.ErrorURI); err != nil {
			writeHTTPError(w, http.StatusBadRequest, err)
			return
		}
	}

	// validate payload
	if _, err := json.Marshal(req.Payload); err != nil {
		writeHTTPError(w, http.StatusBadRequest, err)
		return
	}

	job := types.Job{
		ErrorURI:  req.ErrorURI,
		ExecuteAt: req.ExecuteAt,
		Payload:   req.Payload,
		URI:       req.URI,
	}

	// add job to queue
	if err := h.Job.Enqueue(&job); err != nil {
		writeHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	writeHTTPResponse(w, http.StatusCreated, &job)
}

// ListJobs returns a list of all the jobs in the system
func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.DB.GetJobs()
	if err != nil {
		writeHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	writeHTTPResponse(w, http.StatusOK, jobs)
}
