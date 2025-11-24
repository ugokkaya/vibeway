package router

import (
	"vibeway/internal/config"
	"vibeway/internal/middleware"
	"vibeway/internal/proxy"
	"vibeway/internal/upstream"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App, cfg config.Config, upstreams *upstream.Manager) {
	proxyClient := proxy.NewProxyClient(
		time.Duration(cfg.Server.RequestTimeoutMs)*time.Millisecond,
		time.Duration(cfg.Server.RequestTimeoutMs)*time.Millisecond,
	)

	for _, rCfg := range cfg.Routes {
		// Build middleware chain
		var handlers []any

		// Security first
		handlers = append(handlers, middleware.Security())

		for _, mw := range rCfg.Middlewares {
			switch mw {
			case "jwt":
				handlers = append(handlers, middleware.JWT(cfg.Security.JWT))
			case "ratelimit":
				handlers = append(handlers, middleware.RateLimit(
					cfg.Security.RateLimit.PerRoute,
					time.Minute,
				))
			case "rbac":
				// This needs per-route config for roles.
				// For now, we assume a generic RBAC or we need to extend config.
				// Let's assume the route config might have "allowed_roles" in a real app.
				// Here we just add the middleware placeholder.
				handlers = append(handlers, middleware.RBAC(nil))
			}
		}

		// Proxy handler
		handlers = append(handlers, func(c fiber.Ctx) error {
			u, ok := upstreams.GetUpstream(rCfg.Upstream)
			if !ok {
				return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Upstream not found"})
			}

			targetURL, ok := u.GetNextURL()
			if !ok {
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "No healthy upstream available"})
			}

			// Rewrite path if needed, or just append
			// Simple append: targetURL + c.Path() (if targetURL is base)
			// But targetURL might be full path.
			// Let's assume targetURL is "http://host:port" and we append request URI.

			// Handle path rewriting
			// If route path ends with /*, strip the prefix
			reqPath := c.Path()
			routePath := rCfg.Path
			if strings.HasSuffix(routePath, "/*") {
				prefix := strings.TrimSuffix(routePath, "/*")
				if strings.HasPrefix(reqPath, prefix) {
					reqPath = strings.TrimPrefix(reqPath, prefix)
				}
			}

			// Ensure leading slash
			if !strings.HasPrefix(reqPath, "/") {
				reqPath = "/" + reqPath
			}

			reqURL := targetURL + reqPath

			return proxyClient.Do(c.Request(), c.Response(), reqURL)
		})

		// Register for each method
		// Register for methods
		app.Add(rCfg.Methods, rCfg.Path, handlers[0], handlers[1:]...)
	}
}
