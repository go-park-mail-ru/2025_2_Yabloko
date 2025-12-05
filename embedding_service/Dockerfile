FROM golang:1.25.1-alpine AS builder

RUN apk add --no-cache \
    git \
    build-base \
    cmake \
    gcc \
    g++ \
    make \
    alpine-sdk

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -o /embedding_service_bin \
    ./cmd/embedding_service/main.go

FROM alpine:latest

RUN apk add --no-cache libc6-compat libstdc++

WORKDIR /

COPY --from=builder /embedding_service_bin /embedding_service_bin
RUN chmod +x /embedding_service_bin

RUN mkdir -p /models

VOLUME ["/models"]

ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

EXPOSE 50051

ENTRYPOINT ["/embedding_service_bin"]
