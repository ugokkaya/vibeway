package middleware

import (
	"fmt"
	"strconv"
	"time"

	"vibeway/pkg/cache"
	"vibeway/pkg/logger"

	"github.com/gofiber/fiber/v3"
)

func RateLimit(limit int, window time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		ip := c.IP()
		key := fmt.Sprintf("ratelimit:%s", ip)

		// Simple counter implementation for demonstration
		// In production, use a sliding window Lua script

		countStr, err := cache.Get(c.Context(), key)
		if err != nil && err.Error() != "redis: nil" {
			logger.Error("Rate limit cache error", err, nil)
			// Fail open or closed? Let's fail open for resilience
			return c.Next()
		}

		count := 0
		if countStr != "" {
			count, _ = strconv.Atoi(countStr)
		}

		if count >= limit {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		}

		// Increment and set expiry if new
		if count == 0 {
			cache.Set(c.Context(), key, 1, window)
		} else {
			// Just increment, keeping original TTL would require Lua or separate TTL check
			// For simplicity, we just increment here.
			// Ideally: INCR key
			cache.Client.Incr(c.Context(), key)
		}

		return c.Next()
	}
}
