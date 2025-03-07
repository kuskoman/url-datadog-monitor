FROM golang:1.24.1-alpine3.21 as build

ARG VERSION \
    GIT_COMMIT \
    GITHUB_REPO="github.com/kuskoman/url-datadog-monitor"

WORKDIR /app

# Install SSL certificates
RUN apk add --no-cache ca-certificates && update-ca-certificates

RUN grep "nobody:x:65534" /etc/passwd > /app/user

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s \
    -X ${GITHUB_REPO}/pkg/version.Version=${VERSION} \
    -X ${GITHUB_REPO}/pkg/version.GitCommit=${GIT_COMMIT} \
    -X ${GITHUB_REPO}/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o main cmd/operator/main.go

FROM scratch as release

COPY --from=build /app/user /etc/passwd
COPY --from=build /app/main /app/main
# Copy CA certificates for TLS verification
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8125 8080 8081

USER 65534

ENTRYPOINT ["/app/main"]