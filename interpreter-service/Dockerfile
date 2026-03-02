FROM golang:1.25-alpine AS builder

WORKDIR /workspace

RUN apk add --no-cache ca-certificates
RUN apk add --no-cache build-base

COPY go.work go.work.sum ./
COPY cst-to-ast-service ./cst-to-ast-service
COPY semantic-analyzer-service ./semantic-analyzer-service
COPY interpreter-service ./interpreter-service

WORKDIR /workspace/interpreter-service

RUN go mod download

RUN go build -o /out/interpreter-service ./cmd/main.go

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates && adduser -D -g '' appuser

COPY --from=builder /out/interpreter-service /app/interpreter-service
COPY interpreter-service/config.yaml /app/config.yaml
COPY interpreter-service/config.docker.yaml /app/config.docker.yaml

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/interpreter-service"]
CMD ["-config", "/app/config.yaml"]
