# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o vibeway ./cmd/vibeway

# Final Stage
FROM gcr.io/distroless/static-debian11

WORKDIR /app

COPY --from=builder /app/vibeway .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./vibeway"]
