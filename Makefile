.PHONY: autopilot

ap: autopilot

autopilot:
	go build -o ${GOPATH}/bin/$@ cli/cmd/main.go