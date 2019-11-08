.PHONY: ap

# Build the AutoPilot CLI
ap:
	go generate ./...
	go build -o ${GOPATH}/bin/$@ cli/cmd/main.go

# Build Test Project
.PHONY: test
test:
	cd test/e2e && ./run_test_project.sh


# Clean test
.PHONY: clean
clean:
	rm -rf test/e2e/test