package main

import "github.com/solo-io/autopilot/codegen"

func main() {
	if err := codegen.Run("examples/promota/"); err != nil {
		panic(err)
	}
}
