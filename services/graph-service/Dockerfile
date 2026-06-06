FROM golang:1.24-alpine AS builder
WORKDIR /app

ENV GOTOOLCHAIN=local

RUN apk add --no-cache git ca-certificates
COPY services/graph-service/go.mod services/graph-service/go.sum ./
RUN go mod download
COPY services/graph-service/. .
RUN go build -o /usr/local/bin/graph-service ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates curl
COPY --from=builder /usr/local/bin/graph-service /usr/local/bin/graph-service
WORKDIR /app
EXPOSE 9090 9091
ENTRYPOINT ["/usr/local/bin/graph-service"]
CMD ["graph-service"]
