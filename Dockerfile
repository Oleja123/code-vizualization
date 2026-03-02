FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

COPY cst-to-ast-service/ ./cst-to-ast-service/
COPY semantic-analyzer-service/ ./semantic-analyzer-service/

WORKDIR /app/semantic-analyzer-service

RUN go build -o semantic-analyzer ./cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/semantic-analyzer-service/semantic-analyzer .
COPY --from=builder /app/semantic-analyzer-service/config.yaml .

EXPOSE 8082
CMD ["./semantic-analyzer"]