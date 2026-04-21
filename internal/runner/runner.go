package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"autocheck-microservices/internal/contracts"
)

func RunAllTests(parentCtx context.Context, code string, timeout time.Duration, testCases []contracts.TestCase) []contracts.TestRunResult {
	results := make([]contracts.TestRunResult, len(testCases))
	var wg sync.WaitGroup
	for i, tc := range testCases {
		wg.Add(1)
		go func(index int, testCase contracts.TestCase) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(parentCtx, timeout)
			defer cancel()
			results[index] = runOne(ctx, code, testCase)
		}(i, tc)
	}
	wg.Wait()
	return results
}

func runOne(ctx context.Context, code string, testCase contracts.TestCase) contracts.TestRunResult {
	cmd := exec.CommandContext(ctx, "python3", "-u", "-c", code)
	cmd.Stdin = strings.NewReader(testCase.Input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := contracts.TestRunResult{
		Output:    stdout.String(),
		Error:     "",
		Status:    contracts.TestStatusOK,
		TimeMs:    duration.Milliseconds(),
		TestNum:   testCase.Num,
		InputFile: fmt.Sprintf("input%d.txt", testCase.Num),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.Status = contracts.TestStatusTLE
		result.Error = "Time Limit Exceeded"
		return result
	}

	if err != nil {
		result.Status = contracts.TestStatusRE
		result.Error = strings.TrimSpace(stderr.String())
		if result.Error == "" {
			result.Error = err.Error()
		}
		return result
	}

	if strings.TrimSpace(stdout.String()) != strings.TrimSpace(testCase.Output) {
		result.Status = contracts.TestStatusWA
		result.Error = fmt.Sprintf("Expected: %s\nGot: %s", strings.TrimSpace(testCase.Output), strings.TrimSpace(stdout.String()))
	}

	return result
}

func BuildSummary(results []contracts.TestRunResult) contracts.RunSummary {
	summary := contracts.RunSummary{Total: len(results)}
	for _, item := range results {
		switch item.Status {
		case contracts.TestStatusOK:
			summary.Ok++
		case contracts.TestStatusWA:
			summary.Wa++
		case contracts.TestStatusRE:
			summary.Re++
		case contracts.TestStatusTLE:
			summary.Tle++
		default:
			summary.Pending++
		}
	}
	return summary
}
