package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"autocheck-microservices/internal/contracts"
	"autocheck-microservices/internal/queue"
	"autocheck-microservices/internal/redisstate"
	"autocheck-microservices/internal/runner"
	"autocheck-microservices/internal/storage"
	"autocheck-microservices/internal/testloader"
	"autocheck-microservices/pkg/config"

	"github.com/redis/go-redis/v9"
)

func RunWorker(ctx context.Context, cfg config.Config) error {
	store, err := storage.New(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOUseSSL)
	if err != nil {
		return err
	}
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()
	q := queue.New(redisClient, cfg.QueueName, cfg.EventsChannel)
	stateRepo := redisstate.New(redisClient, cfg.StatePrefix)

	log.Printf("worker started")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		payload, err := q.DequeueJob(ctx, 2*time.Second)
		if err != nil {
			if err == redis.Nil || err == context.DeadlineExceeded {
				continue
			}
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("dequeue job: %v", err)
			continue
		}
		if err = processJob(ctx, store, q, stateRepo, cfg, payload); err != nil {
			log.Printf("process job: %v", err)
		}
	}
}

func processJob(ctx context.Context, store *storage.Store, q *queue.RedisQueue, stateRepo *redisstate.Repository, cfg config.Config, payload []byte) error {
	var job contracts.QueueJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return err
	}
	state, err := stateRepo.Get(ctx, job.RunID)
	if err == nil {
		state.Status = contracts.RunStatusRunning
		state.Summary = &contracts.RunSummary{Pending: 1}
		_ = stateRepo.Save(ctx, state)
	}

	var submission struct {
		Code string `json:"code"`
	}
	if err = store.MustGetJSON(ctx, job.SubmissionBucket, job.SubmissionKey, &submission); err != nil {
		return publishFailure(ctx, q, job, "failed to read submission: "+err.Error())
	}
	cases, err := testloader.LoadTaskTests(ctx, store, cfg.TestsBucket, job.Task)
	if err != nil {
		return publishFailure(ctx, q, job, "failed to load tests: "+err.Error())
	}
	results := runner.RunAllTests(ctx, submission.Code, contracts.DefaultTimeoutMs(job.TimeoutMs), cases)
	summary := runner.BuildSummary(results)
	status := contracts.RunStatusDone
	if summary.Re > 0 || summary.Tle > 0 {
		status = contracts.RunStatusDone
	}
	resultsKey := fmt.Sprintf("%s/result.json", job.RunID)
	finishedAt := time.Now().UTC().Format(time.RFC3339)
	event := contracts.RunCompletedEvent{
		RunID:         job.RunID,
		Task:          job.Task,
		Status:        status,
		FinishedAt:    finishedAt,
		ResultsKey:    resultsKey,
		ResultsBucket: cfg.ResultsBucket,
		Results:       results,
		Summary:       summary,
	}
	if err = store.PutJSON(ctx, cfg.ResultsBucket, resultsKey, event); err != nil {
		return err
	}
	eventPayload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return q.PublishEvent(ctx, eventPayload)
}

func publishFailure(ctx context.Context, q *queue.RedisQueue, job contracts.QueueJob, message string) error {
	event := contracts.RunCompletedEvent{
		RunID:      job.RunID,
		Task:       job.Task,
		Status:     contracts.RunStatusFailed,
		FinishedAt: time.Now().UTC().Format(time.RFC3339),
		Summary:    contracts.RunSummary{Total: 0, Pending: 0},
		Error:      message,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return q.PublishEvent(ctx, payload)
}
