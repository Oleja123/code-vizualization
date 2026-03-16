FROM golang:1.25.6-alpine AS builder

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

COPY cppcheck-analyzer-service/ ./cppcheck-analyzer-service/

WORKDIR /app/cppcheck-analyzer-service

RUN go build -o cppcheck-analyzer ./cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates cppcheck
WORKDIR /app

COPY --from=builder /app/cppcheck-analyzer-service/cppcheck-analyzer .
COPY --from=builder /app/cppcheck-analyzer-service/config.docker.yaml .

EXPOSE 8086
CMD ["./cppcheck-analyzer", "-config", "config.docker.yaml"]
