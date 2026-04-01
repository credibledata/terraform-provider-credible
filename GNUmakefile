default: build

build:
	go build -o terraform-provider-credible

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/credibledata/credible/0.1.0/$(shell go env GOOS)_$(shell go env GOARCH)
	cp terraform-provider-credible ~/.terraform.d/plugins/registry.terraform.io/credibledata/credible/0.1.0/$(shell go env GOOS)_$(shell go env GOARCH)/

test:
	go test ./... -v

vet:
	go vet ./...

fmt:
	gofmt -s -w .

.PHONY: default build install test vet fmt
