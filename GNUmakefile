VERSION  ?= 0.1.0
BINARY    = terraform-provider-credible
OS_ARCH   = $(shell go env GOOS)_$(shell go env GOARCH)
PLUGIN_DIR = $(HOME)/.terraform.d/plugins
# The provider is published to both registries, so a local build is mirrored
# under both namespaces to match whichever source address a config declares.
HOSTS     = registry.terraform.io registry.opentofu.org

default: build

build:
	go build -o $(BINARY)

install: build
	@for host in $(HOSTS); do \
		dir="$(PLUGIN_DIR)/$$host/credibledata/credible/$(VERSION)/$(OS_ARCH)"; \
		mkdir -p "$$dir"; \
		cp $(BINARY) "$$dir/"; \
		echo "installed $$host/credibledata/credible $(VERSION) ($(OS_ARCH))"; \
	done

test:
	go test ./... -v

vet:
	go vet ./...

fmt:
	gofmt -s -w .

.PHONY: default build install test vet fmt
