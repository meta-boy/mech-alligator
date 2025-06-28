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
		INSERT INTO jobs (id, type, status, payload, result, error_message, 
						 attempts, max_attempts, scheduled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	if j.CreatedAt.IsZero() {
		j.CreatedAt = now
	}
	j.UpdatedAt = now

	_, err = r.db.ExecContext(ctx, query,
		j.ID, j.Type, j.Status, payloadJSON, resultJSON,
		j.Error, j.Attempts, j.MaxAttempts, j.ScheduledAt,
		j.CreatedAt, j.UpdatedAt,
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

	j.UpdatedAt = time.Now()

	query := `
		UPDATE jobs 
		SET type = $2, status = $3, payload = $4, result = $5,
			error_message = $6, attempts = $7, max_attempts = $8, 
			scheduled_at = $9, started_at = $10, completed_at = $11, updated_at = $12
		WHERE id = $1
	`

	_, err = r.db.ExecContext(ctx, query,
		j.ID, j.Type, j.Status, payloadJSON, resultJSON,
		j.Error, j.Attempts, j.MaxAttempts, j.ScheduledAt,
		j.StartedAt, j.CompletedAt, j.UpdatedAt,
	)

	return err
}

func (r *JobRepository) GetByID(ctx context.Context, id string) (*job.Job, error) {
	query := `
		SELECT id, type, status, payload, result, error_message,
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
		SELECT id, type, status, payload, result, error_message,
			   attempts, max_attempts, scheduled_at, started_at,
			   completed_at, created_at, updated_at
		FROM jobs
		WHERE status = $1 AND scheduled_at <= CURRENT_TIMESTAMP AND attempts < max_attempts
		ORDER BY scheduled_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	row := r.db.QueryRowContext(ctx, query, job.StatusPending)
	return r.scanJob(row)
}

func (r *JobRepository) ListByStatus(ctx context.Context, status job.Status, limit int) ([]*job.Job, error) {
	query := `
		SELECT id, type, status, payload, result, error_message,
			   attempts, max_attempts, scheduled_at, started_at,
			   completed_at, created_at, updated_at
		FROM jobs
		WHERE status = $1
		ORDER BY scheduled_at DESC
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
	var startedAt, completedAt sql.NullTime
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&j.ID, &j.Type, &j.Status, &payloadJSON, &resultJSON,
		&j.Error, &j.Attempts, &j.MaxAttempts, &j.ScheduledAt,
		&startedAt, &completedAt, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Handle nullable timestamps
	if startedAt.Valid {
		j.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		j.CompletedAt = &completedAt.Time
	}
	j.CreatedAt = createdAt
	j.UpdatedAt = updatedAt

	if err := json.Unmarshal(payloadJSON, &j.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if err := json.Unmarshal(resultJSON, &j.Result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &j, nil
}

func (r *JobRepository) GetNextPendingAndMarkRunning(ctx context.Context) (*job.Job, error) {
	// Use a transaction to ensure atomicity
	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Simple query - just get pending jobs
	query := `
		SELECT id, type, status, payload, result, error_message,
			   attempts, max_attempts, scheduled_at,
			   started_at, completed_at, created_at, updated_at
		FROM jobs
		WHERE status = 'pending' AND attempts < max_attempts
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`

	var j job.Job
	var payloadJSON, resultJSON []byte
	var startedAt, completedAt, createdAt, updatedAt sql.NullTime

	err = tx.QueryRowContext(ctx, query).Scan(
		&j.ID, &j.Type, &j.Status, &payloadJSON, &resultJSON,
		&j.Error, &j.Attempts, &j.MaxAttempts, &j.ScheduledAt,
		&startedAt, &completedAt, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to query job: %w", err)
	}

	// Handle nullable timestamps properly
	if startedAt.Valid {
		j.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		j.CompletedAt = &completedAt.Time
	}
	if createdAt.Valid {
		j.CreatedAt = createdAt.Time
	} else {
		j.CreatedAt = j.ScheduledAt // fallback
	}
	if updatedAt.Valid {
		j.UpdatedAt = updatedAt.Time
	} else {
		j.UpdatedAt = j.ScheduledAt // fallback
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(payloadJSON, &j.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	if err := json.Unmarshal(resultJSON, &j.Result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	// Mark as running
	now := time.Now().UTC()
	j.Status = job.StatusRunning
	j.StartedAt = &now
	j.Attempts++
	j.UpdatedAt = now

	// Update in database
	updateQuery := `
		UPDATE jobs 
		SET status = 'running', started_at = $2, attempts = $3, updated_at = $4
		WHERE id = $1
	`

	_, err = tx.ExecContext(ctx, updateQuery, j.ID, j.StartedAt, j.Attempts, j.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to mark job as running: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &j, nil
}
