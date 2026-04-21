package testloader

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"autocheck-microservices/internal/contracts"
	"autocheck-microservices/internal/storage"
)

func LoadTaskTests(ctx context.Context, store *storage.Store, bucket string, task int32) ([]contracts.TestCase, error) {
	prefix := fmt.Sprintf("task-%d/", task)
	keys, err := store.ListKeys(ctx, bucket, prefix)
	if err != nil {
		return nil, err
	}
	inputKeys := make([]string, 0)
	for _, key := range keys {
		if strings.HasPrefix(filepath.Base(key), "input") && strings.HasSuffix(key, ".txt") {
			inputKeys = append(inputKeys, key)
		}
	}
	sort.Strings(inputKeys)
	cases := make([]contracts.TestCase, 0, len(inputKeys))
	for _, inputKey := range inputKeys {
		var num int
		_, err = fmt.Sscanf(filepath.Base(inputKey), "input%d.txt", &num)
		if err != nil {
			continue
		}
		inputData, err := store.GetObject(ctx, bucket, inputKey)
		if err != nil {
			return nil, err
		}
		outputKey := filepath.Join(prefix, fmt.Sprintf("output%d.txt", num))
		outputData, _ := store.GetObject(ctx, bucket, outputKey)
		cases = append(cases, contracts.TestCase{Num: num, Input: string(inputData), Output: string(outputData)})
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("no tests found for task %d", task)
	}
	return cases, nil
}
