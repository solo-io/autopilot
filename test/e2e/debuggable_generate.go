// use this file to manually invoke code generation (without ap CLI)
// can be useful for debugging
package main

import (
	"log"

	"github.com/solo-io/autopilot/codegen"
)

func main() {
	if err := codegen.Run("canary", false); err != nil {
		log.Fatal(err)
	}
}
