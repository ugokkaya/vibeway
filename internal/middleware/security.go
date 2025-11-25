package middleware

import (
	"regexp"

	"github.com/gofiber/fiber/v3"
)

var (
	// Basic SQL Injection patterns
	// This is NOT a full WAF, but covers common obvious attacks
	sqliPattern = regexp.MustCompile(`(?i)(UNION\s+SELECT|DROP\s+TABLE|INSERT\s+INTO|DELETE\s+FROM|UPDATE\s+\w+\s+SET|--|;\s*$)`)

	// Basic XSS patterns
	xssPattern = regexp.MustCompile(`(?i)(<script|javascript:|on\w+\s*=)`)
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
	if sqliPattern.MatchString(s) {
		return true
	}
	if xssPattern.MatchString(s) {
		return true
	}
	return false
}
