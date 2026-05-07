FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache ca-certificates tzdata

COPY go.work go.work.sum ./
COPY pkg/common ./pkg/common
COPY services/retrieval ./services/retrieval

WORKDIR /workspace/services/retrieval
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/retrieval ./cmd

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/retrieval /app/retrieval

EXPOSE 50051

CMD ["/app/retrieval"]
