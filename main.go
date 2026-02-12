package main

import (
	"fmt"
	"os"

	"github.com/darkstorage/cli/cmd"
)

var Version = "dev"

func main() {
	cmd.SetVersion(Version)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
