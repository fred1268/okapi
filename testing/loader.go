package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fred1268/okapi/testing/internal/log"
)

func readJSONDependencies(directory string, requests []*APIRequest) error {
	for _, request := range requests {
		if err := request.validate(); err != nil {
			return fmt.Errorf("invalid test '%s': %w", request.Name, err)
		}
		if strings.HasPrefix(request.Payload, "@") {
			file := request.Payload[1:]
			if request.Payload == "@file" {
				file = fmt.Sprintf("%s.payload.json", strings.ToLower(request.Name))
			}
			content, err := os.ReadFile(path.Join(directory, file))
			if err != nil {
				file = fmt.Sprintf("payload/%s.json", strings.ToLower(request.Name))
				content, err = os.ReadFile(path.Join(directory, file))
				if err != nil {
					return fmt.Errorf("cannot read test file '%s': %w", file, err)
				}
			}
			request.Payload = string(content)
		}
		if strings.HasPrefix(request.Expected.Response, "@") {
			file := request.Expected.Response[1:]
			if request.Expected.Response == "@file" {
				file = fmt.Sprintf("%s.expected.json", strings.ToLower(request.Name))
			}
			content, err := os.ReadFile(path.Join(directory, file))
			if err != nil {
				file = fmt.Sprintf("expected/%s.json", strings.ToLower(request.Name))
				content, err = os.ReadFile(path.Join(directory, file))
				if err != nil {
					return fmt.Errorf("cannot read test file '%s': %w", file, err)
				}
			}
			request.Expected.Response = string(content)
		}
	}
	return nil
}

func loadTest(cfg *Config, uniqueTests map[string]*APIRequest, filename string) ([]*APIRequest, error) {
	content, err := os.ReadFile(path.Join(cfg.Directory, filename))
	if err != nil {
		return nil, fmt.Errorf("cannot read test file '%s': %w", filename, err)
	}
	var tests struct {
		Tests []*APIRequest
	}
	if err = json.NewDecoder(bytes.NewReader(content)).Decode(&tests); err != nil {
		return nil, fmt.Errorf("cannot decode json file '%s': %w", filename, err)
	}
	for _, test := range tests.Tests {
		if test.Payload == "@file" {
			test.atFile = true
		}
		if test.Expected.Response == "@file" {
			test.Expected.atFile = true
		}
	}
	for _, test := range tests.Tests {
		t, ok := uniqueTests[test.Name]
		if !ok {
			uniqueTests[test.Name] = test
			continue
		}
		log.Printf("Warning: two tests with the same name (%s)\n", test.Name)
		if t.hasFileDepencies() {
			if test.hasFileDepencies() {
				log.Printf("Potential conflict: two tests with the same name (%s) are using @file\n", test.Name)
			}
		} else {
			// replace test without @file with this one
			// doesn't matter if it has @file or not
			uniqueTests[test.Name] = test
		}
	}
	if err := readJSONDependencies(cfg.Directory, tests.Tests); err != nil {
		return nil, err
	}
	return tests.Tests, nil
}

// LoadTests reads all test files in the provided directory and
// returns them sorted by file.
//
// The result is a map indexed by the file name, its value being an
// array of *APIRequests corresponding to the tests in the file.
func LoadTests(cfg *Config) (map[string][]*APIRequest, error) {
	files, err := os.ReadDir(cfg.Directory)
	if err != nil {
		return nil, err
	}
	uniqueTests := make(map[string]*APIRequest)
	allTests := make(map[string][]*APIRequest)
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".test.json") {
			continue
		}
		if file.Name() == "setup.test.json" || file.Name() == "teardown.test.json" {
			continue
		}
		if cfg.File != "" && cfg.File != file.Name() {
			continue
		}
		tests, err := loadTest(cfg, uniqueTests, file.Name())
		if err != nil {
			return nil, err
		}
		if len(tests) == 0 {
			log.Printf("Skipping '%s': no tests found in file\n", file.Name())
			continue
		}
		if cfg.Test != "" {
			var test *APIRequest
			for _, t := range tests {
				if cfg.Test == t.Name {
					test = t
					break
				}
			}
			if test != nil {
				allTests[file.Name()] = []*APIRequest{test}
				break
			}
			continue
		}
		allTests[file.Name()] = tests
	}
	return allTests, nil
}

func readServersConfigs(filename string) (map[string]*ServerConfig, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("file '%s' does not exist: %w", filename, err)
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open file '%s': %w", filename, err)
	}
	var config map[string]*ServerConfig
	if err = json.NewDecoder(bytes.NewReader(content)).Decode(&config); err != nil {
		return nil, fmt.Errorf("cannot decode file '%s': %w", filename, err)
	}
	return config, nil
}

// LoadClients reads the configuration file, creates a map of *Client,
// connects them to their respective servers, and returns it.
//
// If something goes wrong during the JSON parsing or HTTP connection,
// an error will be provided and the map will be nil. The clients can
// then be used to run tests.
func LoadClients(ctx context.Context, cfg *Config) (map[string]*Client, error) {
	serverConfigs, err := readServersConfigs(cfg.Servers)
	if err != nil {
		return nil, err
	}
	clients := make(map[string]*Client)
	for key, value := range serverConfigs {
		client := NewClient(value)
		if err := client.config.validate(); err != nil {
			return nil, fmt.Errorf("server %s: invalid configuration: %w", key, err)
		}
		if client.config.Auth != nil && client.config.Auth.Login != nil {
			if apiResponse, err := client.Connect(ctx); err != nil {
				return nil, fmt.Errorf("cannot connect to server '%s' (response: %v): %w", key, apiResponse, err)
			}
			if cfg.Verbose {
				log.Printf("Connected to %s (%s)\n", key, value.Host)
			}
		}
		clients[key] = client
	}
	return clients, nil
}
