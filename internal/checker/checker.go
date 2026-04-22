package checker

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"myproject/internal/monitor"
)

const (
	numWorkers    = 5
	scheduleEvery = 10 * time.Second
	probeTimeout  = 10 * time.Second
)

type Checker struct {
	repo monitor.Repository
}

func New(repo monitor.Repository) *Checker {
	return &Checker{repo: repo}
}

func (c *Checker) Start(ctx context.Context) {
	jobs := make(chan *monitor.Monitor, numWorkers*2)
	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go c.worker(ctx, i, jobs, &wg)
	}

	c.scheduler(ctx, jobs)
	close(jobs)
	wg.Wait()
	slog.Info("checker: all workers stopped")
}

func (c *Checker) scheduler(ctx context.Context, jobs chan<- *monitor.Monitor) {
	ticker := time.NewTicker(scheduleEvery)
	defer ticker.Stop()
	c.dispatch(ctx, jobs)
	for {
		select {
		case <-ticker.C:
			c.dispatch(ctx, jobs)
		case <-ctx.Done():
			return
		}
	}
}

func (c *Checker) dispatch(ctx context.Context, jobs chan<- *monitor.Monitor) {
	monitors, err := c.repo.ListDueMonitors(ctx)
	if err != nil {
		slog.Error("checker: list due monitors", "err", err)
		return
	}
	for _, m := range monitors {
		select {
		case jobs <- m:
		case <-ctx.Done():
			return
		}
	}
}

func (c *Checker) worker(ctx context.Context, id int,
	jobs <-chan *monitor.Monitor, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case m, ok := <-jobs:
			if !ok {
				return
			}
			c.runCheck(ctx, m)
		case <-ctx.Done():
			return
		}
	}
}

func (c *Checker) runCheck(ctx context.Context, m *monitor.Monitor) {
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	start := time.Now()
	statusCode, errStr := doProbe(probeCtx, m.URL)
	ms := time.Since(start).Milliseconds()

	slog.Info("checker: probed",
		"monitor", m.Name, "status", statusCode, "ms", ms)

	check := &monitor.Check{
		MonitorID:  m.ID,
		StatusCode: statusCode,
		ResponseMs: int(ms),
	}
	if errStr != "" {
		check.Error = &errStr
	}

	saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer saveCancel()

	if err := c.repo.SaveCheck(saveCtx, check); err != nil {
		slog.Error("checker: save check", "err", err)
	}

	now := time.Now()
	m.LastCheckedAt = &now
	if err := c.repo.Update(saveCtx, m); err != nil {
		slog.Error("checker: update last_checked_at", "err", err)
	}
}

func doProbe(ctx context.Context, url string) (int, string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Sprintf("request error: %s", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	return resp.StatusCode, ""
}