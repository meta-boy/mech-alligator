package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/meta-boy/mech-alligator/internal/database"
	"github.com/meta-boy/mech-alligator/internal/domain/job"
)

type JobRepository struct {
	db *database.DB
}

func NewJobRepository(db *database.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(ctx context.Context, j *job.Job) error {
	payloadJSON, err := json.Marshal(j.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resultJSON, err := json.Marshal(j.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	query := `
		INSERT INTO jobs (id, type, priority, status, payload, result, error, 
						 attempts, max_attempts, scheduled_at, started_at, 
						 completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err = r.db.ExecContext(ctx, query,
		j.ID, j.Type, j.Priority, j.Status, payloadJSON, resultJSON,
		j.Error, j.Attempts, j.MaxAttempts, j.ScheduledAt,
		j.StartedAt, j.CompletedAt, j.CreatedAt, j.UpdatedAt,
	)

	return err
}

func (r *JobRepository) Update(ctx context.Context, j *job.Job) error {
	payloadJSON, err := json.Marshal(j.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resultJSON, err := json.Marshal(j.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	query := `
		UPDATE jobs 
		SET type = $2, priority = $3, status = $4, payload = $5, result = $6,
			error = $7, attempts = $8, max_attempts = $9, scheduled_at = $10,
			started_at = $11, completed_at = $12, updated_at = $13
		WHERE id = $1
	`

	j.UpdatedAt = time.Now()

	_, err = r.db.ExecContext(ctx, query,
		j.ID, j.Type, j.Priority, j.Status, payloadJSON, resultJSON,
		j.Error, j.Attempts, j.MaxAttempts, j.ScheduledAt,
		j.StartedAt, j.CompletedAt, j.UpdatedAt,
	)

	return err
}

func (r *JobRepository) GetByID(ctx context.Context, id string) (*job.Job, error) {
	query := `
		SELECT id, type, priority, status, payload, result, error,
			   attempts, max_attempts, scheduled_at, started_at,
			   completed_at, created_at, updated_at
		FROM jobs
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanJob(row)
}

func (r *JobRepository) GetNextPending(ctx context.Context) (*job.Job, error) {
	query := `
		SELECT id, type, priority, status, payload, result, error,
			   attempts, max_attempts, scheduled_at, started_at,
			   completed_at, created_at, updated_at
		FROM jobs
		WHERE status = $1 AND scheduled_at <= $2 AND attempts < max_attempts
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	row := r.db.QueryRowContext(ctx, query, job.StatusPending, time.Now())
	return r.scanJob(row)
}

func (r *JobRepository) ListByStatus(ctx context.Context, status job.Status, limit int) ([]*job.Job, error) {
	query := `
		SELECT id, type, priority, status, payload, result, error,
			   attempts, max_attempts, scheduled_at, started_at,
			   completed_at, created_at, updated_at
		FROM jobs
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*job.Job
	for rows.Next() {
		j, err := r.scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	return jobs, rows.Err()
}

func (r *JobRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM jobs WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *JobRepository) scanJob(scanner interface {
	Scan(dest ...interface{}) error
}) (*job.Job, error) {
	var j job.Job
	var payloadJSON, resultJSON []byte

	err := scanner.Scan(
		&j.ID, &j.Type, &j.Priority, &j.Status, &payloadJSON, &resultJSON,
		&j.Error, &j.Attempts, &j.MaxAttempts, &j.ScheduledAt,
		&j.StartedAt, &j.CompletedAt, &j.CreatedAt, &j.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(payloadJSON, &j.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if err := json.Unmarshal(resultJSON, &j.Result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &j, nil
}