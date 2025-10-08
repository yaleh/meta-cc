package main

import (
	"github.com/yaleh/meta-cc/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
