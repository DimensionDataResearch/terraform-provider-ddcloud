PROVIDER_NAME = ddcloud

VERSION = 1.3.4
VERSION_INFO_FILE = ./$(PROVIDER_NAME)/version-info.go

BIN_DIRECTORY   = _bin
EXECUTABLE_NAME = terraform-provider-$(PROVIDER_NAME)
DIST_ZIP_PREFIX = $(EXECUTABLE_NAME).v$(VERSION)

REPO_BASE     = github.com/DimensionDataResearch
REPO_ROOT     = $(REPO_BASE)/dd-cloud-compute-terraform
PROVIDER_ROOT = $(REPO_ROOT)/$(PROVIDER_NAME)
VENDOR_ROOT   = $(REPO_ROOT)/vendor

default: fmt build test

fmt:
	go fmt $(REPO_ROOT)/...

clean:
	rm -rf $(BIN_DIRECTORY) $(VERSION_INFO_FILE)
	go clean $(REPO_ROOT)/...

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
	docker build -t ddresearch/terraform-provider-$(PROVIDER_NAME) .
	docker tag ddresearch/terraform-provider-$(PROVIDER_NAME) ddresearch/terraform-provider-$(PROVIDER_NAME):v${VERSION}

# Build docker image
push-docker: build-docker
	docker push ddresearch/terraform-provider-$(PROVIDER_NAME):latest
	docker push ddresearch/terraform-provider-$(PROVIDER_NAME):v${VERSION}

# Produce archives for a GitHub release.
dist: build
	cd $(BIN_DIRECTORY)/windows-386 && \
		zip -9 ../$(DIST_ZIP_PREFIX).windows-386.zip $(EXECUTABLE_NAME).exe
	cd $(BIN_DIRECTORY)/windows-amd64 && \
		zip -9 ../$(DIST_ZIP_PREFIX).windows-amd64.zip $(EXECUTABLE_NAME).exe
	cd $(BIN_DIRECTORY)/linux-amd64 && \
		zip -9 ../$(DIST_ZIP_PREFIX).linux-amd64.zip $(EXECUTABLE_NAME)
	cd $(BIN_DIRECTORY)/darwin-amd64 && \
		zip -9 ../$(DIST_ZIP_PREFIX).darwin-amd64.zip $(EXECUTABLE_NAME)

test: fmt version testprovider testmodels testmaps testcompute

testcompute: fmt
	go test -v $(VENDOR_ROOT)/$(REPO_BASE)/go-dd-cloud-compute/...

testprovider: fmt version
	go test -v $(PROVIDER_ROOT) -run=Test${TEST}

testmodels: fmt version
	go test -v $(REPO_ROOT)/models -run=Test${TEST}

testmaps: fmt version
	go test -v $(REPO_ROOT)/maps -run=Test${TEST}

testall: fmt version
	go test -v $(REPO_ROOT)/...

# Run acceptance tests (since they're long-running, enable retry).
testacc: fmt version
	rm -f "${PWD}/AccTest.log"
	TF_ACC=1 TF_LOG=DEBUG TF_LOG_PATH="${PWD}/AccTest.log" \
	MCP_EXTENDED_LOGGING=1 \
	MCP_MAX_RETRY=6 MCP_RETRY_DELAY=10 \
		go test -v \
		$(PROVIDER_ROOT) \
		-timeout 120m \
		-run=TestAcc${TEST}

version: $(VERSION_INFO_FILE)

$(VERSION_INFO_FILE): Makefile
	@echo "Update version info: v$(VERSION)"
	@echo "package $(PROVIDER_NAME)" > $(VERSION_INFO_FILE)
	@echo "" >> $(VERSION_INFO_FILE)
	@echo "// ProviderVersion is the current version of the $(PROVIDER_NAME) terraform provider." >> $(VERSION_INFO_FILE)
	@echo "const ProviderVersion = \"v$(VERSION) (`git rev-parse HEAD`)\"" >> $(VERSION_INFO_FILE)
