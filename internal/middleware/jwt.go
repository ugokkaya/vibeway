package middleware

import (
	"fmt"
	"strings"

	"vibeway/internal/config"
	"vibeway/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func JWT(cfg config.JWTConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Authorization header format",
			})
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				// Also support RSA if needed, but for now assuming HS256 based on config secret
				// If public key path is provided, we should load it.
				// For simplicity in this snippet, we use the secret.
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			logger.Warn("Invalid JWT token", map[string]interface{}{"error": err.Error()})
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Validate Issuer and Audience
		if iss, err := claims.GetIssuer(); err != nil || iss != cfg.Issuer {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid issuer"})
		}
		if aud, err := claims.GetAudience(); err != nil || !contains(aud, cfg.Audience) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid audience"})
		}

		// Forward User-ID
		if sub, err := claims.GetSubject(); err == nil {
			c.Request().Header.Set("X-User-Id", sub)
		}

		// Store claims in context for RBAC
		c.Locals("user", token)
		c.Locals("claims", claims)

		return c.Next()
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
