default: build test

fmt:
	go fmt github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

build: fmt
	go build -o _bin/terraform-provider-ddcloud

test: fmt
	go test -v github.com/DimensionDataResearch/dd-cloud-compute-terraform/...

testacc:
	TF_ACC=1 go test -v github.com/DimensionDataResearch/dd-cloud-compute-terraform -timeout 120m
