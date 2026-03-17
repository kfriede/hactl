package main

import (
	"os"

	"github.com/kfriede/hactl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
