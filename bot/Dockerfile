FROM golang:1.25.1-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bot .

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/bot /app/bot

CMD ["/app/bot"]
