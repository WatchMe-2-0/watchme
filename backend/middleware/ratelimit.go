package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"watchme/utils"
)

// RateLimiter provides per-IP rate limiting
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rate     int           // Max requests
	window   time.Duration // Time window
}

type visitor struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter creates a rate limiter allowing 'rate' requests per 'window'
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}

	// Cleanup expired entries every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > rl.window {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

// Middleware returns a Fiber middleware handler for rate limiting
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			rl.visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			rl.mu.Unlock()
			return c.Next()
		}

		// Reset counter if window has passed
		if time.Since(v.lastSeen) > rl.window {
			v.count = 1
			v.lastSeen = time.Now()
			rl.mu.Unlock()
			return c.Next()
		}

		v.count++
		v.lastSeen = time.Now()

		if v.count > rl.rate {
			rl.mu.Unlock()
			return utils.Error(c, fiber.StatusTooManyRequests, "Rate limit exceeded. Try again later.")
		}

		rl.mu.Unlock()
		return c.Next()
	}
}
