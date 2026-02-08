package main

import (
	"os"

	"github.com/dhanush0x96c/blueprint/cmd"
)

func main() {
	code := cmd.Execute()
	os.Exit(code)
}
