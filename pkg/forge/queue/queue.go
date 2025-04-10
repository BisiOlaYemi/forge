package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Job represents a queued job
type Job struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time             `json:"created_at"`
	Attempts  int                   `json:"attempts"`
	MaxRetries int                  `json:"max_retries"`
}

// Handler processes a job
type Handler func(job *Job) error

// Queue manages job processing
type Queue struct {
	client  *redis.Client
	handlers map[string]Handler
	ctx     context.Context
	cancel  context.CancelFunc
}

// New creates a new queue
func New(addr, password string, db int) (*Queue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithCancel(context.Background())

	queue := &Queue{
		client:   client,
		handlers: make(map[string]Handler),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return queue, nil
}

// RegisterHandler registers a job handler
func (q *Queue) RegisterHandler(jobType string, handler Handler) {
	q.handlers[jobType] = handler
}

// Enqueue adds a job to the queue
func (q *Queue) Enqueue(jobType string, data map[string]interface{}, maxRetries int) (*Job, error) {
	job := &Job{
		ID:         generateID(),
		Type:       jobType,
		Data:       data,
		CreatedAt:  time.Now(),
		Attempts:   0,
		MaxRetries: maxRetries,
	}

	data, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("job:%s", job.ID)
	if err := q.client.Set(q.ctx, key, data, 24*time.Hour).Err(); err != nil {
		return nil, err
	}

	if err := q.client.LPush(q.ctx, fmt.Sprintf("queue:%s", jobType), job.ID).Err(); err != nil {
		return nil, err
	}

	return job, nil
}

// Start begins processing jobs
func (q *Queue) Start() {
	go q.processJobs()
}

// Stop stops processing jobs
func (q *Queue) Stop() {
	q.cancel()
}

// processJobs continuously processes jobs from the queue
func (q *Queue) processJobs() {
	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			for jobType, handler := range q.handlers {
				jobID, err := q.client.RPop(q.ctx, fmt.Sprintf("queue:%s", jobType)).Result()
				if err == redis.Nil {
					continue
				}
				if err != nil {
					continue
				}

				// Get job data
				key := fmt.Sprintf("job:%s", jobID)
				data, err := q.client.Get(q.ctx, key).Result()
				if err != nil {
					continue
				}

				var job Job
				if err := json.Unmarshal([]byte(data), &job); err != nil {
					continue
				}

				// Process job
				if err := handler(&job); err != nil {
					job.Attempts++
					if job.Attempts < job.MaxRetries {
						// Requeue job
						data, _ := json.Marshal(job)
						q.client.Set(q.ctx, key, data, 24*time.Hour)
						q.client.LPush(q.ctx, fmt.Sprintf("queue:%s", jobType), job.ID)
					}
				} else {
					// Delete job on success
					q.client.Del(q.ctx, key)
				}
			}
		}
	}
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
} 