package retry

import (
	"context"
	"log"
	"time"
)

type Config struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// Default Config returns sensible defaults: 3retries, 1s initial wait, 10s cap
func DefualtConfig() Config {
	return Config{
		MaxRetries:  3,
		InitialWait: 1 * time.Second,
		MaxWait:     10 * time.Second,
	}
}

// WithBackoff executes operation withe exponential backoff
func WithBackoff(ctx context.Context, cfg Config, operation func() error) error {
	var err error
	wait := cfg.InitialWait

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retry attempt %d/%d %v", attempt, cfg.MaxRetries, wait)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}

			wait *= 2
			if wait > cfg.MaxWait {
				wait = cfg.MaxWait
			}
		}
		if err := operation(); err == nil {
			return nil
		}

		log.Printf("Operation failed (attempt %d%d): %v", attempt+1, cfg.MaxRetries, err)

	}
	return err
}
