package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

func RBAC(requiredRoles []string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// If no roles required, pass
		if len(requiredRoles) == 0 {
			return c.Next()
		}

		claims, ok := c.Locals("claims").(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "No user claims found"})
		}

		userRolesInterface, ok := claims["roles"].([]interface{})
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "User has no roles"})
		}

		userRoles := make(map[string]bool)
		for _, r := range userRolesInterface {
			if roleStr, ok := r.(string); ok {
				userRoles[roleStr] = true
			}
		}

		for _, required := range requiredRoles {
			if userRoles[required] {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}
