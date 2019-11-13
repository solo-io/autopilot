// Run this file to re-generate CLI docs
package main

import (
	"github.com/solo-io/autopilot/cli/pkg/commands"
	"github.com/spf13/cobra/doc"
	"log"
	"os"
)

//go:generate go run generate_cli_docs.go

func main(){
	err := run()
	if err != nil {
		log.Fatalf("cli docs gen failed: %v", err)
	}
}

func run() error {
	if err := os.RemoveAll("cli"); err != nil {
		return err
	}
	if err := os.MkdirAll("cli", 0777); err != nil {
		return err
	}
	cli := commands.AutoPilotCli()
	if err := doc.GenMarkdownTree(cli, "cli"); err != nil {
		return err
	}
	return nil
}