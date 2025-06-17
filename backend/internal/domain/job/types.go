package job

import (
	"context"
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusRunning    Status = "running"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 2
	PriorityHigh   Priority = 3
	PriorityUrgent Priority = 4
)

type JobType string

const (
	JobTypeScrapeProducts JobType = "scrape_products"
	JobTypeScrapeAllSites JobType = "scrape_all_sites"
)

type Job struct {
	ID          string                 `json:"id" db:"id"`
	Type        JobType                `json:"type" db:"type"`
	Priority    Priority               `json:"priority" db:"priority"`
	Status      Status                 `json:"status" db:"status"`
	Payload     map[string]interface{} `json:"payload" db:"payload"`
	Result      map[string]interface{} `json:"result,omitempty" db:"result"`
	Error       string                 `json:"error,omitempty" db:"error"`
	Attempts    int                    `json:"attempts" db:"attempts"`
	MaxAttempts int                    `json:"max_attempts" db:"max_attempts"`
	ScheduledAt time.Time              `json:"scheduled_at" db:"scheduled_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

type ScrapeJobPayload struct {
	ConfigID     string            `json:"config_id"`
	VendorID     string            `json:"vendor_id"`
	VendorName   string            `json:"vendor_name"`
	SiteURL      string            `json:"site_url"`
	SiteType     string            `json:"site_type"`
	Category     string            `json:"category"`
	Credentials  map[string]string `json:"credentials,omitempty"`
	Options      map[string]string `json:"options,omitempty"`
	AllPages     bool              `json:"all_pages"`
}

type ScrapeJobResult struct {
	ProductsCreated int      `json:"products_created"`
	ProductsUpdated int      `json:"products_updated"`
	ImagesProcessed int      `json:"images_processed"`
	TotalErrors     int      `json:"total_errors"`
	Errors          []string `json:"errors,omitempty"`
	Duration        string   `json:"duration"`
	ScrapedAt       string   `json:"scraped_at"`
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

type Scheduler interface {
	Start(ctx context.Context) error
	Stop() error
	AddJob(job *Job) error
	RemoveJob(id string) error
}