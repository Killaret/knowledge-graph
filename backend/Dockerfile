FROM golang:1.25-alpine AS builder
# Copy entire project (includes knowledge-graph.config.json at root)
COPY . /app
WORKDIR /app/backend
RUN go mod download
# Generate Swagger docs package required by gin-swagger
RUN go run github.com/swaggo/swag/cmd/swag init -g cmd/server/main.go
RUN go build -ldflags="-s -w" -o /app/server ./cmd/server
RUN go build -ldflags="-s -w" -o /app/worker ./cmd/worker
RUN go build -ldflags="-s -w" -o /app/seed ./scripts/seed.go
RUN go build -ldflags="-s -w" -o /app/cli ./cmd/cli

FROM alpine:latest AS base
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

FROM base AS server
COPY --from=builder /app/server .
COPY --from=builder /app/backend/migrations ./migrations
COPY --from=builder /app/backend/openAPI.yaml ./openAPI.yaml
COPY --from=builder /app/backend/docs ./docs
COPY --from=builder /app/knowledge-graph.config.json ./knowledge-graph.config.json
EXPOSE 8080
CMD ["./server"]

FROM base AS worker
COPY --from=builder /app/worker .
COPY --from=builder /app/knowledge-graph.config.json ./knowledge-graph.config.json
CMD ["./worker"]

FROM base AS seed
COPY --from=builder /app/seed .
COPY --from=builder /app/knowledge-graph.config.json ./knowledge-graph.config.json
CMD ["./seed"]

FROM base AS cli
COPY --from=builder /app/cli .
COPY --from=builder /app/knowledge-graph.config.json ./knowledge-graph.config.json
CMD ["./cli"]