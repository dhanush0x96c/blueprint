package main

import (
	"log"

	"github.com/dhanush0x96c/blueprint/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
