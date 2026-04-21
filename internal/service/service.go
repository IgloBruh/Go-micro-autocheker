package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"autocheck-microservices/internal/contracts"
	"autocheck-microservices/internal/queue"
	"autocheck-microservices/internal/redisstate"
	"autocheck-microservices/internal/storage"

	"github.com/google/uuid"
)

type GatewayService struct {
	Store     *storage.Store
	Queue     *queue.RedisQueue
	State     *redisstate.Repository
	SubBucket string
}

func (s *GatewayService) SubmitRun(ctx context.Context, req *contracts.SubmitRunRequest) (*contracts.SubmitRunResponse, error) {
	if req.Code == "" {
		return nil, fmt.Errorf("code is required")
	}
	if req.TimeoutMs <= 0 {
		req.TimeoutMs = 1000
	}

	runID := uuid.NewString()
	createdAt := time.Now().UTC().Format(time.RFC3339)
	submissionKey := fmt.Sprintf("%s/code.py", runID)

	if err := s.Store.PutJSON(ctx, s.SubBucket, submissionKey, map[string]any{
		"run_id":     runID,
		"task":       req.Task,
		"timeout_ms": req.TimeoutMs,
		"code":       req.Code,
		"created_at": createdAt,
	}); err != nil {
		return nil, err
	}

	state := contracts.PersistedRunState{
		Id:        runID,
		Status:    contracts.RunStatusPending,
		Task:      req.Task,
		CreatedAt: createdAt,
		Summary:   &contracts.RunSummary{Pending: 1},
	}
	if err := s.State.Save(ctx, state); err != nil {
		return nil, err
	}

	job := contracts.QueueJob{
		RunID:            runID,
		Task:             req.Task,
		TimeoutMs:        req.TimeoutMs,
		SubmissionKey:    submissionKey,
		SubmissionBucket: s.SubBucket,
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}
	if err = s.Queue.EnqueueJob(ctx, payload); err != nil {
		return nil, err
	}

	return &contracts.SubmitRunResponse{
		Id:        runID,
		Status:    contracts.RunStatusPending,
		Task:      req.Task,
		CreatedAt: createdAt,
	}, nil
}

func (s *GatewayService) GetRun(ctx context.Context, req *contracts.GetRunRequest) (*contracts.GetRunResponse, error) {
	state, err := s.State.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &contracts.GetRunResponse{
		Id:         state.Id,
		Status:     state.Status,
		Task:       state.Task,
		CreatedAt:  state.CreatedAt,
		FinishedAt: state.FinishedAt,
		Results:    state.Results,
		Summary:    state.Summary,
		Error:      state.Error,
	}, nil
}
