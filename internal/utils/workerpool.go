package utils

import (
	"context"
	"os"
	"runtime"
	"sync"
)

type FileResult struct {
	Path    string
	Content []byte
	Matched bool
}

func ProcessFilesParallel(
	ctx context.Context,
	paths []string,
	workers int,
	matcher func(path string, content []byte) bool,
) ([]string, error) {
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	jobs := make(chan string, workers)
	results := make(chan FileResult, workers)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				content, err := os.ReadFile(path)
				if err != nil {
					results <- FileResult{Path: path, Matched: false}
					continue
				}

				matched := matcher(path, content)
				results <- FileResult{Path: path, Matched: matched}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, p := range paths {
			select {
			case <-ctx.Done():
				return
			case jobs <- p:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var matched []string
	for r := range results {
		if r.Matched {
			matched = append(matched, r.Path)
		}
	}

	return matched, nil
}

func ProcessFilesParallelWithContent(
	ctx context.Context,
	paths []string,
	workers int,
	matcher func(path string, content []byte) bool,
) ([]FileResult, error) {
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	jobs := make(chan string, workers)
	results := make(chan FileResult, workers)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				content, err := os.ReadFile(path)
				if err != nil {
					results <- FileResult{Path: path, Matched: false}
					continue
				}

				matched := matcher(path, content)
				if matched {
					results <- FileResult{Path: path, Content: content, Matched: true}
				} else {
					results <- FileResult{Path: path, Matched: false}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, p := range paths {
			select {
			case <-ctx.Done():
				return
			case jobs <- p:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var matched []FileResult
	for r := range results {
		if r.Matched {
			matched = append(matched, r)
		}
	}

	return matched, nil
}
