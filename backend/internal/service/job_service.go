package service

import (
	"context"
	"fmt"
	"time"

	"github.com/meta-boy/mech-alligator/internal/database"
	"github.com/meta-boy/mech-alligator/internal/domain/job"
)

type JobService struct {
	db        *database.DB
	queue     job.Queue
	scheduler job.Scheduler
}

type SiteConfiguration struct {
	ID       string `db:"id"`
	VendorID string `db:"vendor_id"`
	Name     string `db:"name"`
	Endpoint string `db:"endpoint"`
	Type     string `db:"type"`
	Category string `db:"category"`
	Active   bool   `db:"active"`
}

type Vendor struct {
	ID      string `db:"id"`
	Name    string `db:"name"`
	Country string `db:"country"`
}

func NewJobService(db *database.DB, queue job.Queue, scheduler job.Scheduler) *JobService {
	return &JobService{
		db:        db,
		queue:     queue,
		scheduler: scheduler,
	}
}

func (s *JobService) CreateScrapeJob(ctx context.Context, configID string, options map[string]string) (*job.Job, error) {
	// Get site configuration
	config, vendor, err := s.getSiteConfigWithVendor(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to get site configuration: %w", err)
	}

	if !config.Active {
		return nil, fmt.Errorf("site configuration %s is not active", configID)
	}

	// Create job payload
	payload := job.ScrapeJobPayload{
		ConfigID:   config.ID,
		VendorID:   config.VendorID,
		VendorName: vendor.Name,
		SiteURL:    config.Endpoint,
		SiteType:   config.Type,
		Category:   config.Category,
		Options:    options,
		AllPages:   true, // Default to scraping all pages
	}

	// Convert payload to map
	payloadMap := map[string]interface{}{
		"config_id":   payload.ConfigID,
		"vendor_id":   payload.VendorID,
		"vendor_name": payload.VendorName,
		"site_url":    payload.SiteURL,
		"site_type":   payload.SiteType,
		"category":    payload.Category,
		"options":     payload.Options,
		"all_pages":   payload.AllPages,
	}

	// Create job
	j := &job.Job{
		ID:          fmt.Sprintf("scrape_%s_%d", configID, time.Now().Unix()),
		Type:        job.JobTypeScrapeProducts,
		Priority:    job.PriorityNormal,
		Status:      job.StatusPending,
		Payload:     payloadMap,
		MaxAttempts: 3,
		ScheduledAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.queue.Enqueue(ctx, j); err != nil {
		return nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	return j, nil
}

func (s *JobService) CreateScrapeAllSitesJob(ctx context.Context) (*job.Job, error) {
	// Get all active site configurations
	configs, err := s.getActiveSiteConfigurations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active configurations: %w", err)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no active site configurations found")
	}

	// Create individual scrape jobs for each configuration
	var jobIDs []string
	for _, config := range configs {
		scrapeJob, err := s.CreateScrapeJob(ctx, config.ID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create scrape job for config %s: %w", config.ID, err)
		}
		jobIDs = append(jobIDs, scrapeJob.ID)
	}

	// Create parent job to track all scrape jobs
	payload := map[string]interface{}{
		"job_ids":    jobIDs,
		"total_jobs": len(jobIDs),
	}

	j := &job.Job{
		ID:          fmt.Sprintf("scrape_all_%d", time.Now().Unix()),
		Type:        job.JobTypeScrapeAllSites,
		Priority:    job.PriorityNormal,
		Status:      job.StatusCompleted, // Immediately completed since child jobs are queued
		Payload:     payload,
		Result:      map[string]interface{}{"jobs_created": len(jobIDs)},
		MaxAttempts: 1,
		ScheduledAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	now := time.Now()
	j.StartedAt = &now
	j.CompletedAt = &now

	if err := s.queue.Enqueue(ctx, j); err != nil {
		return nil, fmt.Errorf("failed to enqueue parent job: %w", err)
	}

	return j, nil
}

func (s *JobService) GetJob(ctx context.Context, id string) (*job.Job, error) {
	return s.queue.GetJob(ctx, id)
}

func (s *JobService) ListJobs(ctx context.Context, status job.Status, limit int) ([]*job.Job, error) {
	return s.queue.ListJobs(ctx, status, limit)
}

func (s *JobService) CancelJob(ctx context.Context, id string) error {
	j, err := s.queue.GetJob(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	if j == nil {
		return fmt.Errorf("job not found")
	}

	if j.Status == job.StatusRunning {
		return fmt.Errorf("cannot cancel running job")
	}

	if j.Status != job.StatusPending {
		return fmt.Errorf("job is not in pending status")
	}

	j.Status = job.StatusCancelled
	now := time.Now()
	j.CompletedAt = &now

	return s.queue.UpdateJob(ctx, j)
}

func (s *JobService) getSiteConfigWithVendor(ctx context.Context, configID string) (*SiteConfiguration, *Vendor, error) {
	query := `
		SELECT sc.id, sc.vendor_id, sc.name, sc.endpoint, sc.type, sc.category, sc.active,
			   v.id, v.name, v.country
		FROM site_configurations sc
		JOIN vendors v ON sc.vendor_id = v.id
		WHERE sc.id = $1
	`

	row := s.db.QueryRowContext(ctx, query, configID)

	var config SiteConfiguration
	var vendor Vendor

	err := row.Scan(
		&config.ID, &config.VendorID, &config.Name, &config.Endpoint,
		&config.Type, &config.Category, &config.Active,
		&vendor.ID, &vendor.Name, &vendor.Country,
	)

	if err != nil {
		return nil, nil, err
	}

	return &config, &vendor, nil
}

func (s *JobService) getActiveSiteConfigurations(ctx context.Context) ([]SiteConfiguration, error) {
	query := `
		SELECT id, vendor_id, name, endpoint, type, category, active
		FROM site_configurations
		WHERE active = true
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []SiteConfiguration
	for rows.Next() {
		var config SiteConfiguration
		err := rows.Scan(
			&config.ID, &config.VendorID, &config.Name,
			&config.Endpoint, &config.Type, &config.Category, &config.Active,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	return configs, rows.Err()
}