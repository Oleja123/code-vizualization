FROM golang:1.25.6-alpine AS builder

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

COPY go.work go.work.sum ./
COPY cst-to-ast-service/ ./cst-to-ast-service/
COPY cppcheck-analyzer-service/ ./cppcheck-analyzer-service/
COPY semantic-analyzer-service/ ./semantic-analyzer-service/
COPY interpreter-service/ ./interpreter-service/

WORKDIR /app/semantic-analyzer-service

RUN go build -o semantic-analyzer ./cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/semantic-analyzer-service/semantic-analyzer .
COPY --from=builder /app/semantic-analyzer-service/config.yaml .

EXPOSE 8082
CMD ["./semantic-analyzer"]
