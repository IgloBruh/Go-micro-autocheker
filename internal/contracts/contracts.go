package contracts

import "time"

type SubmitRunRequest struct {
	Code      string `json:"code"`
	Task      int32  `json:"task"`
	TimeoutMs int64  `json:"timeout_ms"`
}

type SubmitRunResponse struct {
	Id        string `json:"id"`
	Status    string `json:"status"`
	Task      int32  `json:"task"`
	CreatedAt string `json:"created_at"`
}

type GetRunRequest struct {
	Id string `json:"id"`
}

type GetRunResponse struct {
	Id         string          `json:"id"`
	Status     string          `json:"status"`
	Task       int32           `json:"task"`
	CreatedAt  string          `json:"created_at,omitempty"`
	FinishedAt string          `json:"finished_at,omitempty"`
	Results    []TestRunResult `json:"results,omitempty"`
	Summary    *RunSummary     `json:"summary,omitempty"`
	Error      string          `json:"error,omitempty"`
}

type RunSummary struct {
	Total   int `json:"total"`
	Ok      int `json:"ok"`
	Wa      int `json:"wa"`
	Re      int `json:"re"`
	Tle     int `json:"tle"`
	Pending int `json:"pending"`
}

type QueueJob struct {
	RunID            string `json:"run_id"`
	Task             int32  `json:"task"`
	TimeoutMs        int64  `json:"timeout_ms"`
	SubmissionKey    string `json:"submission_key"`
	SubmissionBucket string `json:"submission_bucket"`
}

type RunCompletedEvent struct {
	RunID         string          `json:"run_id"`
	Task          int32           `json:"task"`
	Status        string          `json:"status"`
	FinishedAt    string          `json:"finished_at"`
	ResultsKey    string          `json:"results_key"`
	ResultsBucket string          `json:"results_bucket"`
	Results       []TestRunResult `json:"results"`
	Summary       RunSummary      `json:"summary"`
	Error         string          `json:"error,omitempty"`
}

type TestCase struct {
	Num    int    `json:"num"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

type TestRunResult struct {
	Output    string `json:"output"`
	Error     string `json:"error"`
	Status    string `json:"status"`
	TimeMs    int64  `json:"time_ms"`
	TestNum   int    `json:"test_num"`
	InputFile string `json:"input_file"`
}

type PersistedRunState struct {
	Id         string          `json:"id"`
	Status     string          `json:"status"`
	Task       int32           `json:"task"`
	CreatedAt  string          `json:"created_at,omitempty"`
	FinishedAt string          `json:"finished_at,omitempty"`
	Summary    *RunSummary     `json:"summary,omitempty"`
	Results    []TestRunResult `json:"results,omitempty"`
	ResultsKey string          `json:"results_key,omitempty"`
	Error      string          `json:"error,omitempty"`
}

const (
	RunStatusPending = "PENDING"
	RunStatusRunning = "RUNNING"
	RunStatusDone    = "DONE"
	RunStatusFailed  = "FAILED"

	TestStatusOK  = "OK"
	TestStatusWA  = "WA"
	TestStatusRE  = "RE"
	TestStatusTLE = "TLE"
)

func DefaultTimeoutMs(v int64) time.Duration {
	if v <= 0 {
		v = 1000
	}
	return time.Duration(v) * time.Millisecond
}
