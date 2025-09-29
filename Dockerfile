FROM golang:1.25
WORKDIR /app

COPY . .
RUN go mod download
RUN go install github.com/jackc/tern/v2@latest
ENV PATH="/go/bin:${PATH}"

RUN chmod +x ./entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]