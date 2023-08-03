package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fred1268/okapi/testing"
)

func help() {
	fmt.Println("okapi is a tool to help make API tests as easy as table driven tests.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println()
	fmt.Println("\tokapi [options] <test_directory>")
	fmt.Println()
	fmt.Println("The options are:")
	fmt.Println()
	fmt.Println("\t--servers-file, -s (mandatory):\t\t\t\tpoint to the configuration file's location")
	fmt.Println("\t--verbose, -v (default no):\t\t\t\tenable verbose mode")
	fmt.Println("\t--file-parallel (default no):\t\t\t\trun the test files in parallel (instead of the tests themselves)")
	fmt.Println("\t--file, -f (default none):\t\t\t\tonly run the specified test file")
	fmt.Println("\t--test, -t (default none):\t\t\t\tonly run the specified standalone test")
	fmt.Println("\t--timeout (default 30s):\t\t\t\tset a default timeout for all HTTP requests")
	fmt.Println("\t--no-parallel (default parallel):\t\t\tprevent tests from running in parallel")
	fmt.Println("\t--user-agent (default okapi UA):\t\t\tset the default user agent")
	fmt.Println("\t--content-type (default 'application/json'):\t\tset the default content type for requests")
	fmt.Println("\t--accept (default 'application/json'):\t\t\tset the default accept header for responses")
	fmt.Println()
	fmt.Println("The parameters are:")
	fmt.Println()
	fmt.Println("\ttest_directory:\t\t\t\t\t\tpoint to the directory where all the test files are located")
	fmt.Println()
	fmt.Println("More information (and source code) on: https://github.com/fred1268/okapi")
	fmt.Println()
}

func main() {
	if len(os.Args) == 1 || len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "help" || os.Args[1] == "-h") {
		help()
		return
	}
	cfg, err := testing.LoadConfig(os.Args)
	if err != nil {
		log.Fatalf("Cannot read command line parameters: %s\n", err)
	}
	if err := testing.Run(context.Background(), cfg); err != nil {
		log.Fatalf("Cannot run tests: %s\n", err)
	}
}
