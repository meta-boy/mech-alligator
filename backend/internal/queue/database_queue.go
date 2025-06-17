package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/meta-boy/mech-alligator/internal/domain/job"
	"github.com/meta-boy/mech-alligator/internal/repository/postgres"
)

type DatabaseQueue struct {
	repo *postgres.JobRepository
}

func NewDatabaseQueue(repo *postgres.JobRepository) *DatabaseQueue {
	return &DatabaseQueue{
		repo: repo,
	}
}

func (q *DatabaseQueue) Enqueue(ctx context.Context, j *job.Job) error {
	if j.ID == "" {
		return fmt.Errorf("job ID is required")
	}

	// Set defaults
	if j.Status == "" {
		j.Status = job.StatusPending
	}
	if j.Priority == 0 {
		j.Priority = job.PriorityNormal
	}
	if j.MaxAttempts == 0 {
		j.MaxAttempts = 3
	}
	if j.ScheduledAt.IsZero() {
		j.ScheduledAt = time.Now()
	}
	if j.CreatedAt.IsZero() {
		j.CreatedAt = time.Now()
	}
	j.UpdatedAt = time.Now()

	return q.repo.Create(ctx, j)
}

func (q *DatabaseQueue) Dequeue(ctx context.Context) (*job.Job, error) {
	j, err := q.repo.GetNextPending(ctx)
	if err != nil {
		return nil, err
	}

	if j == nil {
		return nil, nil // No jobs available
	}

	// Mark as running
	now := time.Now()
	j.Status = job.StatusRunning
	j.StartedAt = &now
	j.Attempts++

	if err := q.repo.Update(ctx, j); err != nil {
		return nil, fmt.Errorf("failed to mark job as running: %w", err)
	}

	return j, nil
}

func (q *DatabaseQueue) UpdateJob(ctx context.Context, j *job.Job) error {
	return q.repo.Update(ctx, j)
}

func (q *DatabaseQueue) GetJob(ctx context.Context, id string) (*job.Job, error) {
	return q.repo.GetByID(ctx, id)
}

func (q *DatabaseQueue) ListJobs(ctx context.Context, status job.Status, limit int) ([]*job.Job, error) {
	return q.repo.ListByStatus(ctx, status, limit)
}

func (q *DatabaseQueue) DeleteJob(ctx context.Context, id string) error {
	return q.repo.Delete(ctx, id)
}