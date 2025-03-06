VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GITHUB_REPO = github.com/kuskoman/url-datadog-monitor
LDFLAGS = -w -s -X $(GITHUB_REPO)/pkg/version.Version=$(VERSION) -X $(GITHUB_REPO)/pkg/version.GitCommit=$(GIT_COMMIT) -X $(GITHUB_REPO)/pkg/version.BuildDate=$(BUILD_DATE)
CGO_ENABLED ?= 0

# Supported platforms
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
BINARIES = url-datadog-monitor-standalone url-datadog-monitor-operator

.PHONY: all
all: build

.PHONY: build
build: build-standalone build-operator

.PHONY: build-standalone
build-standalone:
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o bin/url-datadog-monitor-standalone cmd/standalone/main.go

.PHONY: build-operator
build-operator:
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o bin/url-datadog-monitor-operator cmd/operator/main.go

.PHONY: test
test:
	go test ./... -v

.PHONY: cross-build
cross-build:
	mkdir -p bin/release
	$(foreach platform,$(PLATFORMS),\
		$(foreach binary,$(BINARIES),\
			GOOS=$(firstword $(subst /, ,$(platform))) \
			GOARCH=$(lastword $(subst /, ,$(platform))) \
			CGO_ENABLED=$(CGO_ENABLED) \
			go build -ldflags="$(LDFLAGS)" \
			-o bin/release/$(binary)-$(firstword $(subst /, ,$(platform)))-$(lastword $(subst /, ,$(platform)))$(if $(findstring windows,$(platform)),.exe,) \
			$(if $(findstring standalone,$(binary)),cmd/standalone/main.go,cmd/operator/main.go); \
		)\
	)
	cd bin/release && find . -type f -not -name "*.exe" -exec chmod +x {} \;
	cd bin/release && for f in *; do sha256sum "$$f" > "$$f.sha256"; done

.PHONY: docker-build
docker-build:
	docker build -t ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-standalone-scratch \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f docker/standalone-scratch.Dockerfile .
	
	docker build -t ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-operator-scratch \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f docker/operator-scratch.Dockerfile .
	
	docker build -t ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-standalone-alpine \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f docker/standalone-alpine.Dockerfile .
	
	docker build -t ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-operator-alpine \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-f docker/operator-alpine.Dockerfile .
	
	# Tag latest for each variant
	docker tag ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-standalone-scratch ghcr.io/kuskoman/url-datadog-monitor:latest-standalone-scratch
	docker tag ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-operator-scratch ghcr.io/kuskoman/url-datadog-monitor:latest-operator-scratch
	docker tag ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-standalone-alpine ghcr.io/kuskoman/url-datadog-monitor:latest-standalone-alpine
	docker tag ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-operator-alpine ghcr.io/kuskoman/url-datadog-monitor:latest-operator-alpine
	
	# Tag latest for main variants
	docker tag ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-standalone-scratch ghcr.io/kuskoman/url-datadog-monitor:latest-standalone
	docker tag ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-operator-scratch ghcr.io/kuskoman/url-datadog-monitor:latest-operator
	
	# Tag latest for main image
	docker tag ghcr.io/kuskoman/url-datadog-monitor:$(VERSION)-standalone-scratch ghcr.io/kuskoman/url-datadog-monitor:latest

.PHONY: clean
clean:
	rm -rf bin/