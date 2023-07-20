package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fred1268/okapi/testing/internal/log"
)

type testIn struct {
	file      string
	test      *APIRequest
	client    *Client
	fileStart time.Time
	start     time.Time
	verbose   bool
}

type testOut struct {
	file      string
	fileStart time.Time
	start     time.Time
	fail      bool
	logs      []string
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

func readExpectedJSON(directory string, requests []*APIRequest) error {
	for _, request := range requests {
		if err := request.validate(); err != nil {
			return fmt.Errorf("invalid test '%s': %w", request.Name, err)
		}
		if request.Expected.Response != "@file" {
			continue
		}
		file := fmt.Sprintf("%s.expected.json", strings.ToLower(request.Name))
		content, err := os.ReadFile(path.Join(directory, file))
		if err != nil {
			return fmt.Errorf("cannot read test file '%s': %w", file, err)
		}
		request.Expected.Response = string(content)
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
		if err := readExpectedJSON(directory, tests.Tests); err != nil {
			return nil, err
		}
		allTests[file.Name()] = tests.Tests
	}
	return allTests, nil
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

func runOne(ctx context.Context, tin *testIn, out chan<- *testOut) error {
	tout := &testOut{file: tin.file, fileStart: tin.fileStart, start: tin.start}
	response, err := tin.client.Test(ctx, tin.test, tin.verbose)
	if err != nil {
		if !errors.Is(err, ErrStatusCodeMismatched) && !errors.Is(err, ErrResultMismatched) {
			return fmt.Errorf("cannot run test '%s': %w", tin.file, err)
		}
		tout.fail = true
	}
	tout.logs = append(tout.logs, response.Logs...)
	out <- tout
	return nil
}

// Run starts the tests according to the provided config.
//
// The Config only requires the Servers and Tests values,
// all other fields have reasonable defaults.
func Run(ctx context.Context, cfg *Config) error {
	clients, err := LoadClients(ctx, cfg)
	if err != nil {
		return fmt.Errorf("cannot connect to servers: %w", err)
	}
	allTests, err := LoadTests(cfg.Tests)
	if err != nil {
		return fmt.Errorf("cannot read tests: %w", err)
	}
	out := make(chan *testOut)
	in := make(chan *testIn)
	if cfg.Parallel {
		in = make(chan *testIn, runtime.NumCPU())
	}
	results := make(map[string]struct{})
	var wg sync.WaitGroup
	go func() {
		for {
			run := <-in
			if err := runOne(ctx, run, out); err != nil {
				log.Fatalf("runOne failed: %v\n", err)
				wg.Done()
				return
			}
		}
	}()
	go func() {
		counts := make(map[string]int)
		logs := make(map[string][]string)
		for {
			tout := <-out
			counts[tout.file]++
			if tout.fail {
				results[tout.file] = struct{}{}
			}
			logs[tout.file] = append(logs[tout.file], tout.logs...)
			if counts[tout.file] == len(allTests[tout.file]) {
				lines := logs[tout.file]
				delete(logs, tout.file)
				delete(counts, tout.file)
				if _, ok := results[tout.file]; ok {
					log.Printf("--- FAIL:\t%s\n", tout.file)
				} else if cfg.Verbose {
					log.Printf("--- PASS:\t%s\n", tout.file)
				}
				for _, line := range lines {
					log.Printf(line)
				}
				if _, ok := results[tout.file]; !ok && cfg.Verbose {
					log.Printf("PASS\n")
				}
				if _, ok := results[tout.file]; ok {
					log.Printf("FAIL \n")
					log.Printf("FAIL\t%s\t\t\t%0.3fs\n", tout.file, time.Since(tout.fileStart).Seconds())
					log.Printf("FAIL \n")
				} else {
					log.Printf("ok\t%-30s\t\t\t%0.3fs\n", tout.file, time.Since(tout.fileStart).Seconds())
				}
			}
			wg.Done()
		}
	}()
	for key, tests := range allTests {
		fileStart := time.Now()
		for _, test := range tests {
			if clients[test.Server] == nil {
				log.Fatalf("invalid server for %s ('%s')\n", test.Name, key)
				continue
			}
			tin := &testIn{
				file:      key,
				test:      test,
				client:    clients[test.Server],
				fileStart: fileStart,
				start:     time.Now(),
				verbose:   cfg.Verbose,
			}
			wg.Add(1)
			in <- tin
		}
	}
	wg.Wait()
	close(in)
	close(out)
	return nil
}
