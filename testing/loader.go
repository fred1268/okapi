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
				return fmt.Errorf("cannot read test file '%s': %w", file, err)
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
				return fmt.Errorf("cannot read test file '%s': %w", file, err)
			}
			request.Expected.Response = string(content)
		}
	}
	return nil
}

// LoadTests reads all test files in the provided directory and
// returns them sorted by file.
//
// The result is a map indexed by the file name, its value being an
// array of *APIRequests corresponding to the tests in the file.
func LoadTests(directory string) (map[string][]*APIRequest, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	allTests := make(map[string][]*APIRequest)
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".test.json") {
			continue
		}
		content, err := os.ReadFile(path.Join(directory, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("cannot read test file '%s': %w", file.Name(), err)
		}
		var tests struct {
			Tests []*APIRequest
		}
		if err = json.NewDecoder(bytes.NewReader(content)).Decode(&tests); err != nil {
			return nil, fmt.Errorf("cannot decode json file '%s': %w", file.Name(), err)
		}
		if err := readJSONDependencies(directory, tests.Tests); err != nil {
			return nil, err
		}
		allTests[file.Name()] = tests.Tests
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
