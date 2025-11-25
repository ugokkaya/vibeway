package middleware

import (
	"fmt"
	"time"

	"vibeway/pkg/cache"
	"vibeway/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
)

// Lua script for atomic rate limiting
// KEYS[1]: rate limit key
// ARGV[1]: limit count
// ARGV[2]: window in seconds
var rateLimitScript = redis.NewScript(`
	local key = KEYS[1]
	local limit = tonumber(ARGV[1])
	local window = tonumber(ARGV[2])

	local current = redis.call("GET", key)
	if current and tonumber(current) >= limit then
		return -1
	end

	current = redis.call("INCR", key)
	if tonumber(current) == 1 then
		redis.call("EXPIRE", key, window)
	end

	return current
`)

func RateLimit(limit int, window time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("ratelimit:%s", ip)

		// Execute Lua script
		// Window must be in seconds for EXPIRE
		windowSeconds := int(window.Seconds())
		if windowSeconds < 1 {
			windowSeconds = 1
		}

		result, err := rateLimitScript.Run(c.Context(), cache.Client, []string{key}, limit, windowSeconds).Result()
		if err != nil {
			logger.Error("Rate limit redis error", err, nil)
			// Fail open
			return c.Next()
		}

		// Result is the current count, or -1 if exceeded
		current := result.(int64)

		if current == -1 {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		}

		// Optional: Add headers
		c.Response().Header.Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Response().Header.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", int64(limit)-current))

		return c.Next()
	}
}
