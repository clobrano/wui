package youtube

import (
	"context"
	"log/slog"
	"time"
)

// StartPoller runs Syncer.Sync immediately and then on every interval tick
// until ctx is cancelled. It returns immediately; the loop runs in a goroutine.
func StartPoller(ctx context.Context, syncer *Syncer, interval time.Duration) {
	go func() {
		run := func() {
			result, err := syncer.Sync()
			if err != nil {
				slog.Error("YouTube sync failed", "error", err)
				return
			}
			slog.Info("YouTube sync complete", "added", result.Added, "skipped", result.Skipped)
		}

		run()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				run()
			case <-ctx.Done():
				return
			}
		}
	}()
}
