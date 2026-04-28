FROM golang:1.24-alpine AS builder
# Copy entire project (includes knowledge-graph.config.json at root)
COPY . /app
WORKDIR /app/backend
ENV GOTOOLCHAIN=auto
RUN go mod download
RUN go build -o ../server ./cmd/server
RUN go build -o ../worker ./cmd/worker
RUN go build -o ../seed ./scripts/seed.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/worker .
COPY --from=builder /app/seed .
COPY --from=builder /app/backend/migrations ./migrations
# Copy config file for runtime access
COPY --from=builder /app/knowledge-graph.config.json ./knowledge-graph.config.json
EXPOSE 8080
CMD ["./server"]