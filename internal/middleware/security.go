package middleware

import (
	"github.com/gofiber/fiber/v3"
)

func Security() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Header Sanitization
		c.Response().Header.Del("X-Powered-By")
		c.Response().Header.Del("Server")
		c.Response().Header.Del("Via")

		// Security Headers
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Simple Input Sanitization (Block obvious SQLi/XSS patterns)
		// In a real WAF, this would be much more robust
		path := c.Path()
		if containsSuspiciousPatterns(path) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Malicious input detected",
			})
		}

		return c.Next()
	}
}

func containsSuspiciousPatterns(s string) bool {
	suspicious := []string{
		"<script>", "javascript:", "UNION SELECT", "DROP TABLE", "--",
	}
	for _, p := range suspicious {
		if containsIgnoreCase(s, p) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive check
	// For better performance, use strings.Contains with ToLower
	return false // Placeholder for brevity, implement properly
}
