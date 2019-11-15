// Run this file to re-generate CLI docs
package main

import (
	"log"
	"os"

	"github.com/solo-io/go-utils/clidoc"

	"github.com/solo-io/autopilot/cli/pkg/commands"
)

//go:generate go run generate_cli_docs.go

func main() {
	err := run()
	if err != nil {
		log.Fatalf("cli docs gen failed: %v", err)
	}
}

func run() error {
	if err := os.RemoveAll("content/cli"); err != nil {
		return err
	}
	if err := os.MkdirAll("content/cli", 0777); err != nil {
		return err
	}
	cli := commands.AutopilotCli()
	if err := clidoc.GenerateCliDocsWithConfig(cli, clidoc.Config{OutputDir: "content/cli"}); err != nil {
		return err
	}
	return nil
}
