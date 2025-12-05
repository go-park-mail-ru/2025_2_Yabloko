FROM golang:1.25.1-bookworm AS build-stage

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential libwebp-dev pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN go build -o /embedding_service_bin ./cmd/embedding_service


FROM debian:bookworm-slim

WORKDIR /

COPY --from=build-stage /embedding_service_bin /embedding_service_bin
RUN chmod +x /embedding_service_bin

RUN mkdir -p /models
VOLUME ["/models"]

RUN apt-get update && apt-get install -y --no-install-recommends \
    libc6 libstdc++6 libgomp1 ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY ./lib/libllama_go.so /usr/local/lib/libllama_go.so
RUN chmod +x /usr/local/lib/libllama_go.so

ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

EXPOSE 50051

ENTRYPOINT ["/embedding_service_bin"]
