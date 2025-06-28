package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/meta-boy/mech-alligator/internal/domain/job"
	"github.com/meta-boy/mech-alligator/internal/service"
)

type JobHandler struct {
	jobService *service.JobService
}

func NewJobHandler(jobService *service.JobService) *JobHandler {
	return &JobHandler{
		jobService: jobService,
	}
}

type CreateScrapeJobRequest struct {
	ConfigID string            `json:"config_id"`
	Options  map[string]string `json:"options,omitempty"`
}

type JobResponse struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    int                    `json:"priority"`
	Status      string                 `json:"status"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	ScheduledAt string                 `json:"scheduled_at"`
	StartedAt   *string                `json:"started_at,omitempty"`
	CompletedAt *string                `json:"completed_at,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

func (h *JobHandler) CreateScrapeJob(w http.ResponseWriter, r *http.Request) {
	var req CreateScrapeJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ConfigID == "" {
		http.Error(w, "config_id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	j, err := h.jobService.CreateScrapeJob(ctx, req.ConfigID, req.Options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := h.jobToResponse(j)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *JobHandler) CreateScrapeAllJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	j, err := h.jobService.CreateScrapeAllSitesJob(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := h.jobToResponse(j)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL path (you'll need to implement URL routing)
	jobID := r.URL.Query().Get("id")
	if jobID == "" {
		http.Error(w, "job id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	j, err := h.jobService.GetJob(ctx, jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if j == nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	response := h.jobToResponse(j)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	statusStr := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")

	var status job.Status = job.StatusPending // Default
	if statusStr != "" {
		status = job.Status(statusStr)
	}

	limit := 50 // Default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx := r.Context()
	jobs, err := h.jobService.ListJobs(ctx, status, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var responses []JobResponse
	for _, j := range jobs {
		responses = append(responses, h.jobToResponse(j))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs":  responses,
		"count": len(responses),
	})
}

func (h *JobHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("id")
	if jobID == "" {
		http.Error(w, "job id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.jobService.CancelJob(ctx, jobID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "job cancelled successfully",
	})
}

func (h *JobHandler) jobToResponse(j *job.Job) JobResponse {
	response := JobResponse{
		ID:          j.ID,
		Type:        string(j.Type),
		Status:      string(j.Status),
		Payload:     j.Payload,
		Result:      j.Result,
		Error:       j.Error,
		Attempts:    j.Attempts,
		MaxAttempts: j.MaxAttempts,
		ScheduledAt: j.ScheduledAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:   j.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   j.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if j.StartedAt != nil {
		startedAt := j.StartedAt.Format("2006-01-02T15:04:05Z")
		response.StartedAt = &startedAt
	}

	if j.CompletedAt != nil {
		completedAt := j.CompletedAt.Format("2006-01-02T15:04:05Z")
		response.CompletedAt = &completedAt
	}

	return response
}
