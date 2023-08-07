package testing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fred1268/okapi/testing/internal/log"
	tos "github.com/fred1268/okapi/testing/internal/os"
)

func load(ctx context.Context, cfg *Config, clients map[string]*Client, name string) error {
	if cfg.setupCapture == nil {
		cfg.setupCapture = make(map[string]any)
	}
	uniqueTests := make(map[string]*APIRequest)
	tests, err := loadTest(cfg, uniqueTests, fmt.Sprintf("%s.test.json", name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if cfg.Verbose {
		log.Printf("--- EXEC:\t%s.test.json\n", name)
	}
	for _, test := range tests {
		client := clients[test.Server]
		if client == nil {
			log.Fatalf("    --- FAIL:\tinvalid server '%s' for test '%s'\n", test.Server, test.Name)
			continue
		}
		test.Endpoint = tos.SubstituteCapturedVariable(test.Endpoint, cfg.setupCapture)
		test.Payload = tos.SubstituteCapturedVariable(test.Payload, cfg.setupCapture)
		test.Expected.Response = tos.SubstituteCapturedVariable(test.Expected.Response, cfg.setupCapture)
		response, err := client.Test(ctx, test, cfg.Verbose)
		if err != nil {
			if !errors.Is(err, ErrStatusCodeMismatched) && !errors.Is(err, ErrResponseMismatched) {
				log.Printf("    --- FAIL:\tcannot run %s test '%s': %v\n", name, test.Name, err)
				return err
			}
			log.Printf("    --- FAIL:\tcannot run %s test '%s': %s\n", name, test.Name, err)
		}
		if name == "setup" {
			var r interface{}
			err = json.Unmarshal([]byte(strings.ToLower(response.Response)), &r)
			if err != nil {
				continue
			}
			if obj, ok := r.(map[string]any); ok {
				cfg.setupCapture[test.Name] = obj
			}
		}
	}
	return nil
}

// Setup reads the setup.test.json test file and executes all
// the tests within the file.
//
// Results of these tests are captured into a setup object and
// thus can be accessed using `setup.testname.xxx...`.
func Setup(ctx context.Context, cfg *Config, clients map[string]*Client) error {
	return load(ctx, cfg, clients, "setup")
}

// Teardown reads the teardown.test.json test file and executes
// all the tests within the file.
//
// These tests should revert what has been done in setup in order
// to make the test suite idempotent.
func Teardown(ctx context.Context, cfg *Config, clients map[string]*Client) error {
	return load(ctx, cfg, clients, "teardown")
}
