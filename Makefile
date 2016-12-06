VERSION = 1.1.10
VERSION_INFO_FILE = ./vendor/ddcloud/version-info.go

BIN_DIRECTORY   = _bin
EXECUTABLE_NAME = terraform-provider-ddcloud
DIST_ZIP_PREFIX = $(EXECUTABLE_NAME)-v$(VERSION)

default: fmt build test

fmt:
	go fmt github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

clean:
	rm -rf $(BIN_DIRECTORY) $(VERSION_INFO_FILE)

# Peform a development (current-platform-only) build.
dev: version fmt
	go build -o $(BIN_DIRECTORY)/$(EXECUTABLE_NAME)

# Perform a full (all-platforms) build.
build: version build-windows64 build-windows32 build-linux64 build-mac64

build-windows64: version
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIRECTORY)/windows-amd64/$(EXECUTABLE_NAME).exe

build-windows32: version
	GOOS=windows GOARCH=386 go build -o $(BIN_DIRECTORY)/windows-386/$(EXECUTABLE_NAME).exe

build-linux64: version
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIRECTORY)/linux-amd64/$(EXECUTABLE_NAME)

build-mac64: version
	GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIRECTORY)/darwin-amd64/$(EXECUTABLE_NAME)

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
	cd $(BIN_DIRECTORY)/windows-386 && \
		zip -9 ../$(DIST_ZIP_PREFIX)-windows-386.zip $(EXECUTABLE_NAME).exe
	cd $(BIN_DIRECTORY)/windows-amd64 && \
		zip -9 ../$(DIST_ZIP_PREFIX)-windows-amd64.zip $(EXECUTABLE_NAME).exe
	cd $(BIN_DIRECTORY)/linux-amd64 && \
		zip -9 ../$(DIST_ZIP_PREFIX)-linux-amd64.zip $(EXECUTABLE_NAME)
	cd $(BIN_DIRECTORY)/darwin-amd64 && \
		zip -9 ../$(DIST_ZIP_PREFIX)-darwin-amd64.zip $(EXECUTABLE_NAME)

testall: fmt
	go test -v github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

test: testall
	go test -v github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

testprovider: fmt
	go test -v github.com/DimensionDataResearch/dd-cloud-compute-terraform/vendor/ddcloud

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
