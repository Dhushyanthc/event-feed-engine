FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY db ./db
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /event-feed-engine ./cmd/server

FROM alpine:3.22

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /event-feed-engine /usr/local/bin/event-feed-engine

EXPOSE 8080

USER appuser

ENTRYPOINT ["/usr/local/bin/event-feed-engine"]
