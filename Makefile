.PHONY: ap

#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

OUTDIR:=_output

SOURCES := $(shell find . -name "*.go" | grep -v test.go | grep -v '\.\#*')
RELEASE := "true"
ifeq ($(TAGGED_VERSION),)
	# TAGGED_VERSION := $(shell git describe --tags)
	# This doesn't work in CI, need to find another way...
	TAGGED_VERSION := vdev
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)

LDFLAGS := "-X github.com/solo-io/autopilot/pkg/version.Version=$(VERSION)"
GCFLAGS := all="-N -l"

# Passed by cloudbuild
GCLOUD_PROJECT_ID := $(GCLOUD_PROJECT_ID)
BUILD_ID := $(BUILD_ID)

TEST_IMAGE_TAG := test-$(BUILD_ID)
TEST_ASSET_DIR := $(ROOTDIR)/_test
GCR_REPO_PREFIX := gcr.io/$(GCLOUD_PROJECT_ID)

#----------------------------------------------------------------------------------
# Build
#----------------------------------------------------------------------------------

# Build dependencies
.PHONY: generate-deps
generate-deps:
	go get -u github.com/gobuffalo/packr/packr

# Generated Code & Docs
.PHONY: generated-code
generated-code:
	go generate ./...


# CLI
CLI_DIR=cli

.PHONY: ap
ap: $(OUTDIR)/ap
$(OUTDIR)/ap: $(SOURCES)
	go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go

.PHONY: ap-linux-amd64
ap-linux-amd64: $(OUTDIR)/ap-linux-amd64
$(OUTDIR)/ap-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go

.PHONY: ap-darwin-amd64
ap-darwin-amd64: $(OUTDIR)/ap-darwin-amd64
$(OUTDIR)/ap-darwin-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go

.PHONY: ap-windows-amd64
ap-windows-amd64: $(OUTDIR)/ap-windows-amd64.exe
$(OUTDIR)/ap-windows-amd64.exe: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go


.PHONY: build-cli
build-cli: ap-linux-amd64 ap-darwin-amd64 ap-windows-amd64

.PHONY: install-cli
install-cli:
	go build -o ${GOPATH}/bin/ap $(CLI_DIR)/cmd/main.go


#----------------------------------------------------------------------------------
# Test
#----------------------------------------------------------------------------------

# Build & Run Test Project
.PHONY: test
test:
	cd test/e2e && ./run_test_project.sh


#----------------------------------------------------------------------------------
# Release
#----------------------------------------------------------------------------------

# The code does the proper checking for a TAGGED_VERSION
.PHONY: upload-github-release-assets
upload-github-release-assets: build-cli
	go run ci/upload_github_release_assets.go

#----------------------------------------------------------------------------------
# Clean
#----------------------------------------------------------------------------------

# Important to clean before pushing new releases. Dockerfiles and binaries may not update properly
.PHONY: clean
clean:
	rm -rf codegen/*-packr.go
	rm -rf _output
	rm -rf test/e2e/canary
