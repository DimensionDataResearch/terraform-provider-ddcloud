default: build test

fmt:
	go fmt github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

dev: fmt
	go build -o _bin/terraform-provider-ddcloud

build: build-windows64 build-linux64 build-mac64

build-windows64:
	GOOS=windows GOARCH=amd64 go build -o _bin/windows-amd64/terraform-provider-ddcloud.exe

build-linux64:
	GOOS=linux GOARCH=amd64 go build -o _bin/linux-amd64/terraform-provider-ddcloud

build-mac64:
	GOOS=darwin GOARCH=amd64 go build -o _bin/darwin-amd64/terraform-provider-ddcloud

test: fmt
	go test -v github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

testacc:
	TF_ACC=1 go test -v github.com/DimensionDataResearch/dd-cloud-compute-terraform -timeout 120m
