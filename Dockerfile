FROM golang:1.25 
WORKDIR /app

COPY . .
RUN go mod download
EXPOSE 8080
RUN 
ENTRYPOINT ["go", "run", "."]
