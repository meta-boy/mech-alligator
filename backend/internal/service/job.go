package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/meta-boy/mech-alligator/internal/config"
	"github.com/meta-boy/mech-alligator/internal/database"
	"github.com/meta-boy/mech-alligator/internal/domain/job"
)

type JobService struct {
	db    *database.DB
	queue job.Queue
}

func NewJobService(db *database.DB, queue job.Queue) *JobService {
	return &JobService{
		db:    db,
		queue: queue,
	}
}

func (s *JobService) CreateScrapeJob(ctx context.Context, configID string, options map[string]string) (*job.Job, error) {
	// Get reseller config
	resellerConfig, err := s.getResellerConfig(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reseller config: %w", err)
	}

	if !resellerConfig.Active {
		return nil, fmt.Errorf("reseller config %s is not active", configID)
	}

	// Get reseller info
	reseller, err := s.getReseller(ctx, resellerConfig.ResellerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reseller: %w", err)
	}

	// Determine the appropriate source type
	sourceType := s.determineSourceType(resellerConfig.URL, reseller.Name, resellerConfig.SourceType)

	// Create job payload
	payload := job.ScrapeJobPayload{
		ConfigID:     resellerConfig.ID,
		ResellerID:   reseller.ID,
		ResellerName: reseller.Name,
		URL:          resellerConfig.URL,
		SourceType:   sourceType,
		Category:     resellerConfig.Category,
		Options:      options,
	}

	// Convert to map for storage
	payloadMap := map[string]interface{}{
		"config_id":     payload.ConfigID,
		"reseller_id":   payload.ResellerID,
		"reseller_name": payload.ResellerName,
		"url":           payload.URL,
		"source_type":   payload.SourceType,
		"category":      payload.Category,
		"options":       payload.Options,
	}

	// Create job
	j := &job.Job{
		ID:          fmt.Sprintf("scrape_%s_%d", configID, time.Now().Unix()),
		Type:        job.JobTypeScrapeProducts,
		Status:      job.StatusPending,
		Payload:     payloadMap,
		Result:      make(map[string]interface{}),
		MaxAttempts: 3,
		Attempts:    0,
		ScheduledAt: time.Now().UTC(),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.queue.Enqueue(ctx, j); err != nil {
		return nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	return j, nil
}

// Determine source type based on URL and reseller
func (s *JobService) determineSourceType(url, resellerName, configuredType string) string {
	// If explicitly configured, use that
	if configuredType != "" && configuredType != "AUTO" {
		return configuredType
	}

	// Auto-detect based on URL and reseller
	urlLower := strings.ToLower(url)
	resellerLower := strings.ToLower(resellerName)

	switch {
	case strings.Contains(urlLower, "stackskb.com"):
		return "STACKS"
	case strings.Contains(resellerLower, "stackskb"):
		return "STACKS"
	case strings.Contains(urlLower, ".myshopify.com") || strings.Contains(urlLower, "/products.json"):
		return "SHOPIFY"
	case strings.Contains(urlLower, "/products.json"):
		return "SHOPIFY"
	case strings.Contains(urlLower, "/store/") || strings.Contains(urlLower, "/shop/"):
		return "STACKS"
	default:
		// Default fallback
		return "SHOPIFY"
	}
}

// CreateScrapeAllSitesJob creates individual scrape jobs for all active reseller configs
func (s *JobService) CreateScrapeAllSitesJob(ctx context.Context) (*job.Job, error) {
	// Get all active reseller configs
	configs, err := s.getActiveResellerConfigs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active configs: %w", err)
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("no active reseller configs found")
	}

	// Create individual scrape jobs
	var jobIDs []string
	var failedConfigs []string

	for _, config := range configs {
		job, err := s.CreateScrapeJob(ctx, config.ID, nil)
		if err != nil {
			failedConfigs = append(failedConfigs, fmt.Sprintf("%s: %v", config.ID, err))
			continue
		}
		jobIDs = append(jobIDs, job.ID)
	}

	// Create parent job
	payloadMap := map[string]interface{}{
		"job_ids":        jobIDs,
		"total_jobs":     len(jobIDs),
		"failed_configs": failedConfigs,
		"total_configs":  len(configs),
	}

	result := map[string]interface{}{
		"jobs_created":  len(jobIDs),
		"jobs_failed":   len(failedConfigs),
		"total_configs": len(configs),
	}

	now := time.Now().UTC()
	j := &job.Job{
		ID:          fmt.Sprintf("scrape_all_%d", now.Unix()),
		Type:        job.JobTypeScrapeAllSites,
		Status:      job.StatusCompleted,
		Payload:     payloadMap,
		Result:      result,
		MaxAttempts: 1,
		Attempts:    0,
		ScheduledAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
		StartedAt:   &now,
		CompletedAt: &now,
	}

	if err := s.queue.Enqueue(ctx, j); err != nil {
		return nil, fmt.Errorf("failed to enqueue parent job: %w", err)
	}

	return j, nil
}

// Legacy method for backward compatibility
func (s *JobService) CreateScrapeAllJob(ctx context.Context) (*job.Job, error) {
	return s.CreateScrapeAllSitesJob(ctx)
}

func (s *JobService) GetJob(ctx context.Context, id string) (*job.Job, error) {
	return s.queue.GetJob(ctx, id)
}

func (s *JobService) ListJobs(ctx context.Context, status job.Status, limit int) ([]*job.Job, error) {
	return s.queue.ListJobs(ctx, status, limit)
}

// CancelJob cancels a pending job
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
		return fmt.Errorf("job is not in pending status, current status: %s", j.Status)
	}

	// Update job status to cancelled
	j.Status = job.StatusCancelled
	now := time.Now().UTC()
	j.CompletedAt = &now

	return s.queue.UpdateJob(ctx, j)
}

// RetryJob creates a new job based on a failed job
func (s *JobService) RetryJob(ctx context.Context, id string) (*job.Job, error) {
	originalJob, err := s.queue.GetJob(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get original job: %w", err)
	}

	if originalJob == nil {
		return nil, fmt.Errorf("original job not found")
	}

	if originalJob.Status != job.StatusFailed {
		return nil, fmt.Errorf("can only retry failed jobs, current status: %s", originalJob.Status)
	}

	// Create new job with same payload
	newJob := &job.Job{
		ID:          fmt.Sprintf("retry_%s_%d", id, time.Now().Unix()),
		Type:        originalJob.Type,
		Status:      job.StatusPending,
		Payload:     originalJob.Payload,
		MaxAttempts: originalJob.MaxAttempts,
		ScheduledAt: time.Now(),
	}

	if err := s.queue.Enqueue(ctx, newJob); err != nil {
		return nil, fmt.Errorf("failed to enqueue retry job: %w", err)
	}

	return newJob, nil
}

func (s *JobService) getResellerConfig(ctx context.Context, configID string) (*config.ResellerConfig, error) {
	query := `
		SELECT id, reseller_id, name, url, source_type, category, active, options
		FROM reseller_configs
		WHERE id = $1
	`

	var config config.ResellerConfig
	var optionsJSON []byte

	err := s.db.QueryRowContext(ctx, query, configID).Scan(
		&config.ID, &config.ResellerID, &config.Name, &config.URL,
		&config.SourceType, &config.Category, &config.Active, &optionsJSON,
	)

	if err != nil {
		return nil, err
	}

	// Parse options JSON if not empty
	if len(optionsJSON) > 0 {
		config.Options = make(map[string]string)
		if err := json.Unmarshal(optionsJSON, &config.Options); err != nil {
			// If JSON parsing fails, initialize empty map
			config.Options = make(map[string]string)
		}
	} else {
		config.Options = make(map[string]string)
	}

	return &config, nil
}

func (s *JobService) getReseller(ctx context.Context, resellerID string) (*config.Reseller, error) {
	query := `
		SELECT id, name, country, website, currency, active
		FROM resellers
		WHERE id = $1
	`

	var reseller config.Reseller
	err := s.db.QueryRowContext(ctx, query, resellerID).Scan(
		&reseller.ID, &reseller.Name, &reseller.Country,
		&reseller.Website, &reseller.Currency, &reseller.Active,
	)

	return &reseller, err
}

func (s *JobService) getActiveResellerConfigs(ctx context.Context) ([]config.ResellerConfig, error) {
	query := `
		SELECT id, reseller_id, name, url, source_type, category, active
		FROM reseller_configs
		WHERE active = true
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []config.ResellerConfig
	for rows.Next() {
		var config config.ResellerConfig
		err := rows.Scan(
			&config.ID, &config.ResellerID, &config.Name,
			&config.URL, &config.SourceType, &config.Category, &config.Active,
		)
		if err != nil {
			return nil, err
		}
		// Initialize empty options map for configs retrieved without options
		config.Options = make(map[string]string)
		configs = append(configs, config)
	}

	return configs, rows.Err()
}

// GetJobStats returns statistics about jobs
func (s *JobService) GetJobStats(ctx context.Context) (*JobStats, error) {
	query := `
		SELECT 
			status,
			COUNT(*) as count
		FROM jobs 
		GROUP BY status
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get job stats: %w", err)
	}
	defer rows.Close()

	stats := &JobStats{
		StatusCounts: make(map[string]int),
	}

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.StatusCounts[status] = count
		stats.Total += count
	}

	return stats, rows.Err()
}

type JobStats struct {
	Total        int            `json:"total"`
	StatusCounts map[string]int `json:"status_counts"`
}
