FROM golang:1.24.1-alpine3.21 as build

ARG VERSION \
    GIT_COMMIT \
    GITHUB_REPO="github.com/kuskoman/url-datadog-monitor"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s \
    -X ${GITHUB_REPO}/pkg/version.Version=${VERSION} \
    -X ${GITHUB_REPO}/pkg/version.GitCommit=${GIT_COMMIT} \
    -X ${GITHUB_REPO}/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o main cmd/operator/main.go

FROM alpine:3.21 as release

RUN apk add --no-cache curl busybox-extras

WORKDIR /app
COPY --from=build /app/main /app/main

EXPOSE 8125 8080 8081

ENTRYPOINT ["/app/main"]