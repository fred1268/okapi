package main

import (
	"context"
	"log"
	"os"

	"github.com/fred1268/okapi/testing"
)

func main() {
	cfg, err := testing.LoadConfig(os.Args)
	if err != nil {
		log.Fatalf("Cannot read command line parameters: %s\n", err)
	}
	if err := testing.Run(context.Background(), cfg); err != nil {
		log.Fatalf("Cannot run tests: %s\n", err)
	}
}
