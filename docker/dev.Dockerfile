FROM cosmtrek/air:v1.61.7 as air
FROM golang:1.24.1-alpine3.21

WORKDIR /app

RUN apk add --no-cache curl busybox-extras

COPY --from=air /go/bin/air /go/bin/air

EXPOSE 8125 8080 8081

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/go/bin/air", "-c", ".air.toml"]