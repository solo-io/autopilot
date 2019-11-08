.PHONY: autopilot

autopilot: ap

# Build the AutoPilot CLI
ap:
	go generate ./...
	go build -o ${GOPATH}/bin/$@ cli/cmd/main.go

.PHONY: test
test:
	cd test/e2e && ./gen-project.sh
