FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o url-datadog-exporter ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/url-datadog-exporter .
COPY config.yaml .

CMD ["./url-datadog-exporter"]