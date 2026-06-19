package main

import (
	"os"

	"github.com/bensheeler/send/cli"
)

func main() {
	cmd := cli.NewRootCommand(os.Stdout, os.Stderr)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
