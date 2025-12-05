FROM golang:1.25.1-bookworm AS build-stage

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential libwebp-dev pkg-config cmake git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

WORKDIR /tmp/llama-build
RUN git clone --recurse-submodules https://github.com/kelindar/search.git . && \
    mkdir build && cd build && \
    cmake -DBUILD_SHARED_LIBS=ON -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_CXX_COMPILER=g++ -DCMAKE_C_COMPILER=gcc .. && \
    cmake --build . --config Release

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

COPY ./lib/libllama_go.so /usr/local/lib/libllama_go.so
RUN chmod +x /usr/local/lib/libllama_go.so

RUN mkdir -p /models

VOLUME ["/models"]

RUN apt-get update && apt-get install -y --no-install-recommends \
    libgomp1 libc6 libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

EXPOSE 50051

ENTRYPOINT ["/embedding_service_bin"]
