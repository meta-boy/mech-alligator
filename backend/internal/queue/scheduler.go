package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/meta-boy/mech-alligator/internal/domain/job"
)

type JobScheduler struct {
	queue    job.Queue
	handlers map[job.JobType]job.Handler
	workers  int
	stopCh   chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

func NewJobScheduler(queue job.Queue, workers int) *JobScheduler {
	return &JobScheduler{
		queue:    queue,
		handlers: make(map[job.JobType]job.Handler),
		workers:  workers,
		stopCh:   make(chan struct{}),
	}
}

func (s *JobScheduler) RegisterHandler(handler job.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[handler.GetType()] = handler
}

func (s *JobScheduler) Start(ctx context.Context) error {
	log.Printf("Starting job scheduler with %d workers", s.workers)

	for i := 0; i < s.workers; i++ {
		s.wg.Add(1)
		go s.worker(ctx, i)
	}

	// Start cleanup routine
	s.wg.Add(1)
	go s.cleanupWorker(ctx)

	return nil
}

func (s *JobScheduler) Stop() error {
	log.Println("Stopping job scheduler...")
	close(s.stopCh)
	s.wg.Wait()
	log.Println("Job scheduler stopped")
	return nil
}

func (s *JobScheduler) AddJob(j *job.Job) error {
	ctx := context.Background()
	return s.queue.Enqueue(ctx, j)
}

func (s *JobScheduler) RemoveJob(id string) error {
	ctx := context.Background()
	return s.queue.DeleteJob(ctx, id)
}

func (s *JobScheduler) worker(ctx context.Context, workerID int) {
	defer s.wg.Done()

	log.Printf("Worker %d started", workerID)
	defer log.Printf("Worker %d stopped", workerID)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.processNextJob(ctx, workerID)
		}
	}
}

func (s *JobScheduler) processNextJob(ctx context.Context, workerID int) {
	j, err := s.queue.Dequeue(ctx)
	if err != nil {
		log.Printf("Worker %d: Failed to dequeue job: %v", workerID, err)
		return
	}

	if j == nil {
		// No jobs available
		return
	}

	log.Printf("Worker %d: Processing job %s (type: %s)", workerID, j.ID, j.Type)

	s.mu.RLock()
	handler, exists := s.handlers[j.Type]
	s.mu.RUnlock()

	if !exists {
		log.Printf("Worker %d: No handler found for job type %s", workerID, j.Type)
		s.markJobFailed(ctx, j, fmt.Errorf("no handler found for job type %s", j.Type))
		return
	}

	// Create a timeout context for job execution
	jobCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Execute the job
	err = handler.Handle(jobCtx, j)
	if err != nil {
		log.Printf("Worker %d: Job %s failed: %v", workerID, j.ID, err)
		s.markJobFailed(ctx, j, err)
		return
	}

	// Mark job as completed
	now := time.Now()
	j.Status = job.StatusCompleted
	j.CompletedAt = &now

	if err := s.queue.UpdateJob(ctx, j); err != nil {
		log.Printf("Worker %d: Failed to update completed job %s: %v", workerID, j.ID, err)
	} else {
		log.Printf("Worker %d: Job %s completed successfully", workerID, j.ID)
	}

	pendingJobs, err := s.queue.ListJobs(ctx, job.StatusPending, 5)
	if err == nil {
		log.Printf("Worker %d: Found %d pending jobs in database", workerID, len(pendingJobs))
		for _, pj := range pendingJobs {
			log.Printf("Worker %d: Pending job %s, scheduled: %v, attempts: %d/%d",
				workerID, pj.ID, pj.ScheduledAt, pj.Attempts, pj.MaxAttempts)
		}
	}
}

func (s *JobScheduler) markJobFailed(ctx context.Context, j *job.Job, jobErr error) {
	j.Error = jobErr.Error()

	// Check if we should retry
	if j.Attempts < j.MaxAttempts {
		// Retry with exponential backoff
		backoffDuration := time.Duration(j.Attempts*j.Attempts) * time.Minute
		j.ScheduledAt = time.Now().Add(backoffDuration)
		j.Status = job.StatusPending
		j.StartedAt = nil
		log.Printf("Job %s will be retried in %v (attempt %d/%d)",
			j.ID, backoffDuration, j.Attempts, j.MaxAttempts)
	} else {
		// Max attempts reached, mark as failed
		now := time.Now()
		j.Status = job.StatusFailed
		j.CompletedAt = &now
		log.Printf("Job %s failed permanently after %d attempts", j.ID, j.Attempts)
	}

	if err := s.queue.UpdateJob(ctx, j); err != nil {
		log.Printf("Failed to update failed job %s: %v", j.ID, err)
	}
}

func (s *JobScheduler) cleanupWorker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.cleanupOldJobs(ctx)
		}
	}
}

func (s *JobScheduler) cleanupOldJobs(ctx context.Context) {
	// Clean up completed jobs older than 7 days
	cutoff := time.Now().AddDate(0, 0, -7)

	// This would need to be implemented in the repository
	// For now, just log the action
	log.Printf("Cleaning up jobs older than %v", cutoff.Format("2006-01-02"))
}
