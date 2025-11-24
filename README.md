# Vibeway (Go + Fiber)

A high-performance, modular, and secure API Gateway built with **Go 1.25+**, **Fiber v3**, **Redis**, **Prometheus**, and **OpenTelemetry**.

> ⚡️ **Built with Vibe Coding** — Clean, modular, and production-ready.

## 🚀 Features

- **High Performance**: Built on `fasthttp` (via Fiber v3).
- **Dynamic Routing**: Configuration-driven routing with hot reload.
- **Load Balancing**: Round-robin and (placeholder) least-connections strategies.
- **Resilience**: Circuit Breaker, Retries, Timeouts, and Health Checks.
- **Security**:
  - JWT Authentication (HS256/RS256)
  - RBAC (Role-Based Access Control)
  - Rate Limiting (Redis Sliding Window)
  - WAF-like protections (SQLi/XSS blocking, Header sanitization)
- **Observability**:
  - Structured JSON Logging (Zerolog)
  - Metrics (Prometheus)
  - Distributed Tracing (OpenTelemetry/Jaeger)

## 📂 Project Structure

```
api-gateway/
├── cmd/gateway/       # Main entry point
├── internal/
│   ├── config/        # Viper configuration loader
│   ├── router/        # Dynamic route builder
│   ├── proxy/         # Reverse proxy engine
│   ├── middleware/    # Auth, RateLimit, Security
│   ├── upstream/      # Load Balancer, Health Checks, Circuit Breaker
│   ├── metrics/       # Prometheus collectors
│   └── tracing/       # OpenTelemetry setup
├── pkg/               # Shared utilities (Logger, Redis)
├── configs/           # Configuration files
└── docker/            # Dockerfile and Compose
```

## 🛠️ Prerequisites

- Go 1.22+
- Docker & Docker Compose

## 🏃 Quick Start

### 1. Run with Docker Compose (Recommended)

This will start the Gateway, Redis, Prometheus, Jaeger, and a mock User Service.

```bash
docker-compose -f docker/docker-compose.yml up --build
```

### 2. Configuration

Modify `configs/routes.yaml` to define your routes:

```yaml
routes:
  - path: "/api/v1/users/*"
    methods: ["GET", "POST"]
    upstream: "user-service"
    middlewares: ["jwt", "ratelimit"]
```

### 3. Testing

**Health Check:**
```bash
curl http://localhost:8081/health
```

**Proxy Request (requires JWT):**
```bash
curl -H "Authorization: Bearer <TOKEN>" http://localhost:8081/api/v1/users/123
```

**Google Proxy (No Auth):**
```bash
curl -L http://localhost:8081/google/
```

**Metrics:**
```bash
curl http://localhost:8081/metrics
```

## 🔒 Security

- **JWT**: Ensure `security.jwt.secret` is set via environment variable `SECURITY_JWT_SECRET` in production.
- **Rate Limiting**: Configured in `configs/routes.yaml` and `internal/config`.
- **TLS**: Terminate TLS at the load balancer level (AWS ALB, Nginx) or configure Fiber to listen on TLS.

## 📊 Observability

- **Prometheus**: `http://localhost:9093`
- **Jaeger UI**: `http://localhost:16687`

## ⚠️ Notes

- This project uses **Fiber v3 (Beta)**.
- Ensure your local Go environment is 1.22+ to build locally.
- If running locally without Docker, ensure Redis and Jaeger are accessible.
