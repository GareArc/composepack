package main

import (
	"log"

	"composepack/internal/cli"
	"composepack/internal/di"
)

func main() {
	application, err := di.InitializeApplication()
	if err != nil {
		log.Fatal(err)
	}

	if err := cli.NewRootCommand(application).Execute(); err != nil {
		log.Fatal(err)
	}
}
