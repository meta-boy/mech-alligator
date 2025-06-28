package job

import (
	"context"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type JobType string

const (
	JobTypeScrapeProducts JobType = "scrape_products"
	JobTypeScrapeAllSites JobType = "scrape_all_sites"
	JobTypeTagProduct     JobType = "tag_product"
)

type Job struct {
	ID          string                 `json:"id" db:"id"`
	Type        JobType                `json:"type" db:"type"`
	Status      Status                 `json:"status" db:"status"`
	Payload     map[string]interface{} `json:"payload" db:"payload"`
	Result      map[string]interface{} `json:"result,omitempty" db:"result"`
	Error       string                 `json:"error,omitempty" db:"error_message"`
	Attempts    int                    `json:"attempts" db:"attempts"`
	MaxAttempts int                    `json:"max_attempts" db:"max_attempts"`
	ScheduledAt time.Time              `json:"scheduled_at" db:"scheduled_at"`

	// Optional fields that may not exist in simplified schema
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
}

type Handler interface {
	Handle(ctx context.Context, job *Job) error
	GetType() JobType
}

type Queue interface {
	Enqueue(ctx context.Context, job *Job) error
	Dequeue(ctx context.Context) (*Job, error)
	UpdateJob(ctx context.Context, job *Job) error
	GetJob(ctx context.Context, id string) (*Job, error)
	ListJobs(ctx context.Context, status Status, limit int) ([]*Job, error)
	DeleteJob(ctx context.Context, id string) error
}

type ScrapeJobPayload struct {
	ConfigID     string            `json:"config_id"`
	ResellerID   string            `json:"reseller_id"`
	ResellerName string            `json:"reseller_name"`
	URL          string            `json:"url"`
	SourceType   string            `json:"source_type"`
	Category     string            `json:"category"`
	Options      map[string]string `json:"options,omitempty"`
}
