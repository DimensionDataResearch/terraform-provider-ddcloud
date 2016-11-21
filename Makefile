VERSION = 1.1.5
VERSION_INFO_FILE = ./vendor/ddcloud/version-info.go

default: fmt build test

fmt:
	go fmt github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

# Peform a development (current-platform-only) build.
dev: version fmt
	go build -o _bin/terraform-provider-ddcloud

# Perform a full (all-platforms) build.
build: version build-windows64 build-windows32 build-linux64 build-mac64

build-windows64:
	GOOS=windows GOARCH=amd64 go build -o _bin/windows-amd64/terraform-provider-ddcloud.exe

build-windows32:
	GOOS=windows GOARCH=386 go build -o _bin/windows-386/terraform-provider-ddcloud.exe

build-linux64:
	GOOS=linux GOARCH=amd64 go build -o _bin/linux-amd64/terraform-provider-ddcloud

build-mac64:
	GOOS=darwin GOARCH=amd64 go build -o _bin/darwin-amd64/terraform-provider-ddcloud

# Build docker image
build-docker: build-linux64
	docker build -t ddresearch/terraform-provider-ddcloud .
	docker tag ddresearch/terraform-provider-ddcloud ddresearch/terraform-provider-ddcloud:v${VERSION}

# Build docker image
push-docker: build-docker
	docker push ddresearch/terraform-provider-ddcloud:latest
	docker push ddresearch/terraform-provider-ddcloud:v${VERSION}

# Produce archives for a GitHub release.
dist: build
	cd _bin/windows-386 && \
		zip -9 ../windows-386.zip terraform-provider-ddcloud.exe
	cd _bin/windows-amd64 && \
		zip -9 ../windows-amd64.zip terraform-provider-ddcloud.exe
	cd _bin/linux-amd64 && \
		zip -9 ../linux-amd64.zip terraform-provider-ddcloud
	cd _bin/darwin-amd64 && \
		zip -9 ../darwin-amd64.zip terraform-provider-ddcloud

test: fmt
	go test -v github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

# Run acceptance tests (since they're long-running, enable retry).
testacc: fmt
	rm -f "${PWD}/AccTest.log"
	TF_ACC=1 TF_LOG=DEBUG TF_LOG_PATH="${PWD}/AccTest.log" \
	MCP_EXTENDED_LOGGING=1 \
	MCP_MAX_RETRY=6 MCP_RETRY_DELAY=10 \
		go test -v \
		github.com/DimensionDataResearch/dd-cloud-compute-terraform/vendor/ddcloud \
		-timeout 120m \
		-run=TestAcc${TEST}

version: $(VERSION_INFO_FILE)

$(VERSION_INFO_FILE): Makefile
	@echo "Update version info: v$(VERSION)"
	@echo "package ddcloud\n\n// ProviderVersion is the current version of the ddcloud terraform provider.\nconst ProviderVersion = \"v$(VERSION) (`git rev-parse HEAD`)\"" > $(VERSION_INFO_FILE)
