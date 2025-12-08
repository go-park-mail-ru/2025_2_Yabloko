FROM golang:1.25.1-bookworm AS build-stage

RUN apt-get update && apt-get install -y --no-install-recommends \
    wget ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN wget -q https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-x64-1.23.2.tgz \
    && tar -xzf onnxruntime-linux-x64-1.23.2.tgz \
    && ls -la onnxruntime-linux-x64-1.23.2/lib/ \
    && cp onnxruntime-linux-x64-1.23.2/lib/libonnxruntime.so* /usr/local/lib/ \
    && ldconfig \
    && rm -rf onnxruntime-linux-x64-1.23.2 onnxruntime-linux-x64-1.23.2.tgz

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN go build -o /embedding_service_bin ./cmd/embedding_service

FROM debian:bookworm-slim AS build-release-stage

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libstdc++6 \
    libgomp1 \
    grpcurl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build-stage /usr/local/lib/libonnxruntime.so* /usr/local/lib/
RUN ldconfig

COPY --from=build-stage /embedding_service_bin /embedding_service_bin
RUN chmod +x /embedding_service_bin

RUN mkdir -p /models

VOLUME ["/models"]

ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

EXPOSE 50051

ENTRYPOINT ["/embedding_service_bin"]
