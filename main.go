package main

import (
	"os"

	"github.com/yaleh/meta-cc/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
