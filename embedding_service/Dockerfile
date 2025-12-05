FROM golang:1.25.1-alpine AS build-stage

RUN apk add --no-cache git build-base libwebp-dev musl-dev gcc

WORKDIR /app

ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN go build -o /embedding_service_bin ./cmd/embedding_service


FROM alpine:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /embedding_service_bin /embedding_service_bin
RUN chmod +x /embedding_service_bin

RUN mkdir -p /models

VOLUME ["/models"]

RUN apk add --no-cache libgomp libstdc++

COPY ./lib/libllama_go.so /usr/local/lib/libllama_go.so

ENV LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH

EXPOSE 50051

ENTRYPOINT ["/embedding_service_bin"]
