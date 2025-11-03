FROM golang:1.25.1-alpine AS build-stage

RUN apk add --no-cache git build-base libwebp-dev musl-dev gcc

WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o /auth_service_bin ./cmd/auth_service

FROM alpine:latest AS build-release-stage

WORKDIR /

COPY --from=build-stage /auth_service_bin /auth_service_bin
RUN chmod +x /auth_service_bin

EXPOSE 8082

ENTRYPOINT ["/auth_service_bin"]