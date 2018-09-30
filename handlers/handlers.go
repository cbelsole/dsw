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
	uuid "github.com/satori/go.uuid"
)

type (
	Handler struct {
		DB  *db.DB
		Job processors.Job
	}
	createJobRequest struct {
		ErrorURI  *string           `json:"error_uri"`
		ExecuteAt time.Time         `json:"execute_at"`
		Payload   map[string]string `json:"payload"`
		URI       string            `json:"uri"`
	}
	jobResponse struct {
		ID        uuid.UUID         `json:"id" db:"id"`
		ErrorURI  *string           `json:"error_uri" db:"error_uri"`
		ExecuteAt time.Time         `json:"execute_at" db:"execute_at"`
		Payload   map[string]string `json:"payload" db:"payload"`
		URI       string            `json:"uri" db:"uri"`
		CreatedAt time.Time         `json:"created_at" db:"created_at"`
		UpdatedAt time.Time         `json:"updated_at" db:"updated_at"`
	}
)

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.Ping(); err != nil {
		log.Printf("health check failed: %s\n", err)
		writeHTTPResponse(w, http.StatusInternalServerError, map[string]string{"message": "I'm unhealthy"})
	} else {
		writeHTTPResponse(w, http.StatusOK, map[string]string{"message": "I'm healthy"})
	}
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var req createJobRequest
	if err := decoder.Decode(&req); err != nil {
		writeHTTPError(w, http.StatusBadRequest, err)
		return
	}

	if _, err := url.ParseRequestURI(req.URI); err != nil {
		writeHTTPError(w, http.StatusBadRequest, err)
		return
	}

	if req.ErrorURI != nil {
		if _, err := url.ParseRequestURI(*req.ErrorURI); err != nil {
			writeHTTPError(w, http.StatusBadRequest, err)
			return
		}
	}

	payload, err := json.Marshal(req.Payload)
	if err != nil {
		writeHTTPError(w, http.StatusBadRequest, err)
		return
	}

	job := types.Job{
		ErrorURI:  req.ErrorURI,
		ExecuteAt: req.ExecuteAt,
		Payload:   payload,
		URI:       req.URI,
	}

	if err := h.DB.CreateJob(&job); err != nil {
		writeHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.Job.Enqueue(&job); err != nil {
		writeHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	resp, err := newJobResponse(&job)
	if err != nil {
		writeHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	writeHTTPResponse(w, http.StatusCreated, resp)
}

func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.DB.GetJobs()
	if err != nil {
		writeHTTPError(w, http.StatusInternalServerError, err)
		return
	}

	resp := make([]*jobResponse, 0, len(jobs))
	for _, job := range jobs {
		j, err := newJobResponse(job)
		if err != nil {
			writeHTTPError(w, http.StatusInternalServerError, err)
			return
		}

		resp = append(resp, j)
	}

	writeHTTPResponse(w, http.StatusOK, resp)
}

func newJobResponse(job *types.Job) (*jobResponse, error) {
	var payload map[string]string
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return nil, err
	}

	return &jobResponse{
		ID:        job.ID,
		ErrorURI:  job.ErrorURI,
		ExecuteAt: job.ExecuteAt,
		Payload:   payload,
		URI:       job.URI,
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}, nil
}
