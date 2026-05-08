FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache ca-certificates tzdata

COPY go.work go.work.sum ./
COPY pkg/common ./pkg/common
COPY services/auth ./services/auth
COPY services/gateway ./services/gateway
COPY services/document ./services/document
COPY services/ingestion-worker ./services/ingestion-worker
COPY services/retrieval ./services/retrieval

WORKDIR /workspace/services/gateway
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/gateway ./cmd

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/gateway /app/gateway

EXPOSE 8080

CMD ["/app/gateway"]
