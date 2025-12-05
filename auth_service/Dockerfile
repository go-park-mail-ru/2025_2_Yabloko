FROM golang:1.25.1-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    git \
    pkg-config \
    libssl-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -v \
    -o /embedding_service_bin \
    ./cmd/embedding_service/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    libstdc++6 \
    libgomp1 \
    libc6 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /

COPY --from=builder /embedding_service_bin /embedding_service_bin
RUN chmod +x /embedding_service_bin

RUN mkdir -p /models

VOLUME ["/models"]

EXPOSE 50051

ENTRYPOINT ["/embedding_service_bin"]
