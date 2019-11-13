.PHONY: ap

OUTDIR:=_output

# codegen dependencies
codegen-deps:
	go get -v

# Build the AutoPilot CLI
generated-code:
	go generate ./...

ap:
	go build -o $(OUTDIR)/$@ cli/cmd/main.go

install-ap:
	go build -o ${GOPATH}/bin/$@ cli/cmd/main.go

# Cross-platform
ap-linux-amd64:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ cli/cmd/main.go


# Build Test Project
.PHONY: test
test:
	cd test/e2e && ./run_test_project.sh


# Clean test
.PHONY: clean
clean:
	rm -rf test/e2e/test