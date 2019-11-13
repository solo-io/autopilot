// use this file to manually invoke code generation (without ap CLI)
// can be useful for debugging
package main

import (
	"github.com/solo-io/autopilot/codegen"
	"log"
)

func main() {
	if err := codegen.Run("canary", false); err != nil {
		log.Fatal(err)
	}
}