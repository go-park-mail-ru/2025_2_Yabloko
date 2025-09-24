FROM golang:1.25 AS builder
WORKDIR /app

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download


COPY . .
RUN go build -trimpath -ldflags="-s -w" -o app .


FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -S app \
    && adduser -S -G app -H -s /sbin/nologin app

COPY --from=builder --chown=app:app /app/app /app/app
USER app
EXPOSE 8080
ENTRYPOINT ["/app/app"]
