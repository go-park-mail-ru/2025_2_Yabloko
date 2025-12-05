FROM golang:1.25.1-alpine AS build-stage

RUN apk add --no-cache git build-base libwebp-dev musl-dev gcc

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o /embedding_service_bin ./cmd/embedding_service

#FROM build-stage AS run-test-stage
#RUN go test -v ./...

FROM alpine:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /embedding_service_bin /embedding_service_bin
RUN chmod +x /embedding_service_bin

RUN mkdir -p /models

VOLUME ["/models"]

COPY ./lib/libllama_go.so /usr/local/lib/libllama_go.so

EXPOSE 50051

ENTRYPOINT ["/embedding_service_bin"]