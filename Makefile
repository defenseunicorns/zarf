# Figure out which Zarf binary we should use based on the operating system we are on
ZARF_BIN := ./build/zarf
UNAME_S := $(shell uname -s)
UNAME_P := $(shell uname -p)
# Provide a default value for the operating system architecture used in tests, e.g. " TESTDISTRO=provided make test-e2e ARCH=arm64"
ARCH ?= amd64
ifneq ($(UNAME_S),Linux)
	ifeq ($(UNAME_S),Darwin)
		ZARF_BIN := $(addsuffix -mac,$(ZARF_BIN))
	endif
	ifeq ($(UNAME_P),i386)
		ZARF_BIN := $(addsuffix -intel,$(ZARF_BIN))
	endif
	ifeq ($(UNAME_P),arm)
		ZARF_BIN := $(addsuffix -apple,$(ZARF_BIN))
	endif
endif

CLI_VERSION := $(if $(shell git describe --tags), $(shell git describe --tags), "UnknownVersion")
BUILD_ARGS := -s -w -X 'github.com/defenseunicorns/zarf/src/config.CLIVersion=$(CLI_VERSION)'
.DEFAULT_GOAL := help

.PHONY: help
help: ## Show a list of all targets
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sed -n 's/^\(.*\): \(.*\)##\(.*\)/\1:\3/p' \
	| column -t -s ":"

remove-packages: ## remove all zarf packages recursively
	find . -type f -name 'zarf-package-*' -delete

vm-init: ## usage -> make vm-init OS=ubuntu
	vagrant destroy -f
	vagrant up --no-color ${OS}
	echo -e "\n\n\n\033[1;93m  ✅ BUILD COMPLETE.  To access this environment, run \"vagrant ssh ${OS}\"\n\n\n"

vm-destroy: ## Destroy the VM
	vagrant destroy -f

clean: ## Clean the build dir
	rm -rf build

build-cli-linux-amd: build-injector-registry
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(BUILD_ARGS)" -o build/zarf src/main.go

build-cli-linux-arm: build-injector-registry
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(BUILD_ARGS)" -o build/zarf-arm src/main.go

build-cli-mac-intel: build-injector-registry
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(BUILD_ARGS)" -o build/zarf-mac-intel src/main.go

build-cli-mac-apple: build-injector-registry
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(BUILD_ARGS)" -o build/zarf-mac-apple src/main.go

build-cli-linux: build-cli-linux-amd build-cli-linux-arm 

build-cli: build-cli-linux-amd build-cli-linux-arm build-cli-mac-intel build-cli-mac-apple ## Build the CLI

build-injector-registry:
	cd src/injector/stage2 && $(MAKE) build-bootstrap-registry

# Inject and deploy a new dev version of zarf agent for testing (should have an existing zarf agent deployemt)
# @todo: find a clean way to support Kind or k3d: k3d image import $(tag)
dev-agent-image:
	$(eval tag := defenseunicorns/dev-zarf-agent:$(shell date +%s))
	$(eval arch := $(shell uname -m))
	CGO_ENABLED=0 GOOS=linux go build -o build/zarf-linux-$(arch) src/main.go
	DOCKER_BUILDKIT=1 docker build --tag $(tag) --build-arg TARGETARCH=$(arch) . && \
	kind load docker-image zarf-agent:$(tag) && \
	kubectl -n zarf set image deployment/agent-hook server=$(tag)

init-package: ## Create the zarf init package, macos "brew install coreutils" first
	$(ZARF_BIN) package create --confirm --architecture amd64
	$(ZARF_BIN) package create --confirm --architecture arm64
	mv zarf-init-*.tar.zst build
	cd build && sha256sum -b zarf* > zarf.sha256
	ls -lh build

build-test: build-cli init-package ## Build the CLI and create the init package

ci-release: init-package ## Create the init package

.PHONY: package-example-game
package-example-game: ## Create the Doom example
	cd examples/game && ../../$(ZARF_BIN) package create --confirm && mv zarf-package-* ../../build/

.PHONY: package-example-component-scripts
package-example-component-scripts: ## Create component script example
	cd examples/component-scripts && ../../$(ZARF_BIN) package create --confirm && mv zarf-package-* ../../build/

.PHONY: package-example-component-choice
package-example-component-choice: ## Create component choice example
	cd examples/component-choice && ../../$(ZARF_BIN) package create --confirm && mv zarf-package-* ../../build/

.PHONY: package-example-component-variables
package-example-component-variables: ## Create component script example
	cd examples/component-variables && ../../$(ZARF_BIN) package create --confirm && mv zarf-package-* ../../build/

.PHONY: package-example-data-injection
package-example-data-injection: ## create the Zarf package for the data injection example
	cd examples/data-injection && ../../$(ZARF_BIN) package create --confirm && mv zarf-package-* ../../build/

.PHONY: package-example-single-big-bang-package
package-example-single-big-bang-package: ## Create the Zarf package for single-big-bang-package example
	cd examples/single-big-bang-package && ../../$(ZARF_BIN) package create --confirm && mv zarf-package-* ../../build/

.PHONY: package-example-gitops-data
package-example-gitops-data:
	cd examples/gitops-data && ../../$(ZARF_BIN) package create --confirm && mv zarf-package-* ../../build/

.PHONY: package-example-tiny-kafka
package-example-tiny-kafka:
	cd examples/tiny-kafka && ../../$(ZARF_BIN) package create --confirm && mv zarf-package-* ../../build/

.PHONY: package-example-compose
package-example-compose:
	cd examples/composable-packages && ../../$(ZARF_BIN) package create --confirm && mv zarf-package-* ../../build/

# TODO: This can be cleaned up a little more when `zarf init` is able to provide the path to the `zarf-init-<arch>.tar.zst`
.PHONY: test-e2e
test-e2e: ## Run e2e tests. Will automatically build any required dependencies that aren't present. Requires env var TESTDISTRO=[provided|kind|k3d|k3s]
	@#Check to make sure all the packages we need exist
	@if [ ! -f $(ZARF_BIN) ]; then\
		$(MAKE) build-cli;\
	fi
	@if [ ! -f ./build/zarf-init-$(ARCH).tar.zst ]; then\
		$(MAKE) init-package;\
	fi
	@if [ ! -f ./build/zarf-package-appliance-demo-multi-games-$(ARCH).tar.zst ]; then\
		$(MAKE) package-example-game;\
	fi
	@if [ ! -f zarf-package-component-scripts-$(ARCH).tar.zst ]; then\
		$(MAKE) package-example-component-scripts;\
	fi
	@if [ ! -f zarf-package-component-choice-$(ARCH).tar.zst ]; then\
		$(MAKE) package-example-component-choice;\
	fi
	@if [ ! -f zarf-package-component-variables-$(ARCH).tar.zst ]; then\
		$(MAKE) package-example-component-variables;\
	fi
	@if [ ! -f ./build/zarf-package-data-injection-demo-$(ARCH).tar ]; then\
		$(MAKE) package-example-data-injection;\
	fi
	@if [ ! -f ./build/zarf-package-gitops-service-data-$(ARCH).tar.zst ]; then\
		$(MAKE) package-example-gitops-data;\
	fi
	@if [ ! -f ./build/zarf-package-compose-example-$(ARCH).tar.zst ]; then\
		$(MAKE) package-example-compose;\
	fi
	cd src/test/e2e && cp ../../../build/zarf-init-$(ARCH).tar.zst . && go test ./... -v -count=1 -timeout 2400s && rm zarf-init-$(ARCH).tar.zst
