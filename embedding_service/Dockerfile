FROM golang:1.25.1-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -o /embedding_service_bin \
    ./cmd/embedding_service/main.go

FROM alpine:latest

WORKDIR /

COPY --from=builder /embedding_service_bin /embedding_service_bin
RUN chmod +x /embedding_service_bin

EXPOSE 50051

ENTRYPOINT ["/embedding_service_bin"]
