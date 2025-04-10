package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Queue represents a Redis-based queue
type Queue struct {
	client   *redis.Client
	handlers map[string]Handler
	ctx      context.Context
	cancel   context.CancelFunc
}

// Config represents the queue configuration
type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// Handler represents a job handler function
type Handler func(job *Job) error

// Job represents a job in the queue
type Job struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Data       map[string]interface{} `json:"data"`
	CreatedAt  time.Time             `json:"created_at"`
	Attempts   int                    `json:"attempts"`
	MaxRetries int                    `json:"max_retries"`
}

// New creates a new queue instance
func New(host string, password string, db int) (*Queue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithCancel(context.Background())

	return &Queue{
		client:   client,
		handlers: make(map[string]Handler),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// RegisterHandler registers a handler for a job type
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

	// Serialize job data
	jobData, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("job:%s", job.ID)
	if err := q.client.Set(q.ctx, key, jobData, 0).Err(); err != nil {
		return nil, err
	}

	// Add to queue
	if err := q.client.LPush(q.ctx, "queue", job.ID).Err(); err != nil {
		return nil, err
	}

	return job, nil
}

// Start starts processing jobs
func (q *Queue) Start() {
	go q.processJobs()
}

// Stop stops processing jobs
func (q *Queue) Stop() {
	q.cancel()
}

// processJobs processes jobs from the queue
func (q *Queue) processJobs() {
	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			// Get next job from queue
			jobID, err := q.client.RPop(q.ctx, "queue").Result()
			if err != nil {
				if err == redis.Nil {
					time.Sleep(time.Second)
					continue
				}
				continue
			}

			// Get job data
			key := fmt.Sprintf("job:%s", jobID)
			jobData, err := q.client.Get(q.ctx, key).Bytes()
			if err != nil {
				continue
			}

			// Unmarshal job
			var job Job
			if err := json.Unmarshal(jobData, &job); err != nil {
				continue
			}

			// Process job
			if handler, ok := q.handlers[job.Type]; ok {
				if err := handler(&job); err != nil {
					job.Attempts++
					if job.Attempts < job.MaxRetries {
						// Requeue job
						q.client.LPush(q.ctx, "queue", job.ID)
					}
				}
			}

			// Delete job data
			q.client.Del(q.ctx, key)
		}
	}
}

// generateID generates a unique job ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
} 