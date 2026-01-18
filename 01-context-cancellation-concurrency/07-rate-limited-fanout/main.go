package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

type FanOutClient struct {
	client  *http.Client
	limiter *rate.Limiter
	sem     *semaphore.Weighted
}

func NewFanOutClient(limit float64, burst int, maxInFlight int64) *FanOutClient {
	return &FanOutClient{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				MaxIdleConnsPerHost: int(maxInFlight),
			},
			Timeout: 10 * time.Second,
		},
		limiter: rate.NewLimiter(rate.Limit(limit), burst),
		sem:     semaphore.NewWeighted(maxInFlight),
	}
}

func (f *FanOutClient) FetchOne(ctx context.Context, userId int) ([]byte, error) {
	// 并发限制
	if err := f.sem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer f.sem.Release(1)

	// 速率限制
	if err := f.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	if userId > 99999 {
		return nil, fmt.Errorf("fail fast")
	}

	url := "https://api.sampleapis.com/avatar/info"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (f *FanOutClient) FetchAll(ctx context.Context, userIDs []int) (map[int][]byte, error) {

	g, ctx := errgroup.WithContext(ctx)
	results := make(map[int][]byte)
	mu := &sync.Mutex{}

	for _, userId := range userIDs {
		uid := userId
		g.Go(func() error {

			start := time.Now()
			data, err := f.FetchOne(ctx, uid)

			slog.Info("request completed",
				slog.Int("userID", userId),
				slog.Duration("latency", time.Since(start)),
				slog.Bool("success", err == nil))

			if err != nil {
				return err
			}

			mu.Lock()
			results[uid] = data
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

func main() {
	client := NewFanOutClient(10, 8, 20)
	resp, err := client.FetchAll(context.Background(), []int{1, 2, 3, 4, 5})
	if err != nil {
		slog.Error("client fetch all", slog.Any("err", err))
		return
	}
	data, _ := json.Marshal(resp)
	slog.Info("client fetch all", slog.String("resp", string(data)))
}
