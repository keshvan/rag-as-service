FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache ca-certificates tzdata

COPY go.work go.work.sum ./
COPY pkg/common ./pkg/common
COPY services/document ./services/document
COPY services/auth ./services/auth
COPY services/gateway ./services/gateway
COPY services/ingestion-worker ./services/ingestion-worker

WORKDIR /workspace/services/document
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/document ./cmd

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/document /app/document
COPY --from=builder /workspace/services/document/migrations /app/migrations

EXPOSE 50051

CMD ["/app/document"]
