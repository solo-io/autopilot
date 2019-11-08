.PHONY: autopilot

autopilot: ap

# Build the AutoPilot CLI
ap:
	go generate ./...
	go build -o ${GOPATH}/bin/$@ cli/cmd/main.go

