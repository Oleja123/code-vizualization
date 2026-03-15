FROM golang:1.25.6-alpine AS builder

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

# Копируем go.work чтобы Go понял что это воркспейс
COPY go.work go.work.sum ./

# Копируем все модули которые нужны interpreter-service
COPY cst-to-ast-service/ ./cst-to-ast-service/
COPY cppcheck-analyzer-service/ ./cppcheck-analyzer-service/
COPY interpreter-service/ ./interpreter-service/
COPY semantic-analyzer-service/ ./semantic-analyzer-service/

WORKDIR /app/interpreter-service

RUN go build -o interpreter ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/interpreter-service/interpreter .
COPY --from=builder /app/interpreter-service/config.docker.yaml .

EXPOSE 8084
CMD ["./interpreter", "-config", "config.docker.yaml"]
