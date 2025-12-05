FROM golang:1.25.1-alpine AS builder

RUN apk add --no-cache \
    git \
    build-base \
    cmake \
    gcc \
    g++ \
    make \
    alpine-sdk \
    curl \
    wget

WORKDIR /app

RUN mkdir -p /tmp/llama.cpp && cd /tmp/llama.cpp && \
    git clone --depth 1 https://github.com/ggerganov/llama.cpp.git . && \
    mkdir -p /usr/local/include/llama && \
    cp common.h /usr/local/include/llama/ && \
    cp llama.h /usr/local/include/llama/ || true

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    CGO_CFLAGS="-I/usr/local/include/llama" \
    CGO_LDFLAGS="-L/usr/local/lib" \
    go build -o /embedding_service_bin \
    ./cmd/embedding_service/main.go

# Final stage
FROM alpine:latest

RUN apk add --no-cache libc6-compat libstdc++ libgomp

WORKDIR /

COPY --from=builder /embedding_service_bin /embedding_service_bin
RUN chmod +x /embedding_service_bin

RUN mkdir -p /models

VOLUME ["/models"]

EXPOSE 50051

ENTRYPOINT ["/embedding_service_bin"]
