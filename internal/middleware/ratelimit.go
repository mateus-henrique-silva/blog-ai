package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/config"
)

type RateLimiter struct {
	mu      sync.Mutex
	records map[string][]time.Time
	max     int
	window  time.Duration
}

func NewRateLimiter(cfg *config.Config) *RateLimiter {
	rl := &RateLimiter{
		records: make(map[string][]time.Time),
		max:     cfg.RateLimitLogin,
		window:  cfg.RateLimitWindow,
	}
	return rl
}

// Middleware returns a Fiber handler that enforces the rate limit per IP.
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if !rl.allow(ip) {
			retryAfter := int(rl.window.Seconds())
			c.Set("Retry-After", intStr(retryAfter))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "too many attempts — please try again later",
			})
		}
		return c.Next()
	}
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	timestamps := rl.records[ip]
	// Filter within window
	valid := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	valid = append(valid, now)
	rl.records[ip] = valid

	return len(valid) <= rl.max
}

// Cleanup removes stale entries. Call in a background goroutine.
func (rl *RateLimiter) Cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for ip, timestamps := range rl.records {
			valid := timestamps[:0]
			for _, t := range timestamps {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.records, ip)
			} else {
				rl.records[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

func intStr(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
