FROM golang:1.23-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/graph-service ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /usr/local/bin/graph-service /usr/local/bin/graph-service
EXPOSE 9090 9091
ENTRYPOINT ["/usr/local/bin/graph-service"]
