.PHONY: autopilot

ap: autopilot

autopilot:
	go generate ./...
	go build -o ${GOPATH}/bin/$@ cli/cmd/main.go