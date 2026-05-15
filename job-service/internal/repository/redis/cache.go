package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/freelance-platform/job-service/internal/domain"
)

const jobTTL = 60 * time.Second

type JobCache struct {
	client *redis.Client
}

func NewJobCache(client *redis.Client) *JobCache {
	return &JobCache{client: client}
}

func (c *JobCache) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	key := fmt.Sprintf("job:%s", jobID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err // redis.Nil means cache miss
	}
	var job domain.Job
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (c *JobCache) SetJob(ctx context.Context, job *domain.Job) error {
	key := fmt.Sprintf("job:%s", job.ID)
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, jobTTL).Err()
}

func (c *JobCache) DeleteJob(ctx context.Context, jobID string) error {
	return c.client.Del(ctx, fmt.Sprintf("job:%s", jobID)).Err()
}

func (c *JobCache) GetJobList(ctx context.Context, page, pageSize int) ([]*domain.Job, error) {
	key := fmt.Sprintf("jobs:list:%d:%d", page, pageSize)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var jobs []*domain.Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (c *JobCache) SetJobList(ctx context.Context, page, pageSize int, jobs []*domain.Job) error {
	key := fmt.Sprintf("jobs:list:%d:%d", page, pageSize)
	data, err := json.Marshal(jobs)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, jobTTL).Err()
}

func (c *JobCache) InvalidateListCache(ctx context.Context) error {
	// Scan and delete all list keys — called on create/update/delete
	iter := c.client.Scan(ctx, 0, "jobs:list:*", 0).Iterator()
	for iter.Next(ctx) {
		c.client.Del(ctx, iter.Val())
	}
	return iter.Err()
}
