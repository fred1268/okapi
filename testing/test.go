package testing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fred1268/okapi/testing/internal/log"
	"github.com/fred1268/okapi/testing/internal/os"
)

type testIn struct {
	file      string
	test      *APIRequest
	client    *Client
	fileStart time.Time
	start     time.Time
	config    *Config
}

type testOut struct {
	file      string
	fileStart time.Time
	start     time.Time
	fail      bool
	logs      []string
	config    *Config
}

func runOne(ctx context.Context, tin *testIn, out chan<- *testOut) (*APIResponse, error) {
	tout := &testOut{file: tin.file, fileStart: tin.fileStart, start: tin.start, config: tin.config}
	response, err := tin.client.Test(ctx, tin.test, tin.config.Verbose)
	if err != nil {
		if !errors.Is(err, ErrStatusCodeMismatched) && !errors.Is(err, ErrResponseMismatched) {
			tout.fail = true
			tout.logs = append(tout.logs, fmt.Sprintf("    --- FAIL:\tcannot run test '%s' from '%s': %v\n",
				tin.test.Name, tin.file, err))
			out <- tout
			return response, fmt.Errorf("cannot run test '%s' from '%s': %w", tin.test.Name, tin.file, err)
		}
		tout.fail = true
	}
	tout.logs = append(tout.logs, response.Logs...)
	out <- tout
	return response, nil
}

func worker(ctx context.Context, in chan []*testIn, out chan *testOut, done chan bool) {
	for {
		select {
		case runs := <-in:
			captures := make(map[string]any)
			for _, run := range runs {
				run.test.Endpoint = os.SubstituteCapturedVariable(run.test.Endpoint, captures)
				run.test.Payload = os.SubstituteCapturedVariable(run.test.Payload, captures)
				run.test.Expected.Response = os.SubstituteCapturedVariable(run.test.Expected.Response, captures)
				if run.config.FileParallel {
					run.start = time.Now()
				}
				resp, err := runOne(ctx, run, out)
				if err != nil {
					continue
				}
				if run.test.Capture {
					var r interface{}
					err := json.Unmarshal([]byte(strings.ToLower(resp.Response)), &r)
					if err != nil {
						continue
					}
					if obj, ok := r.(map[string]any); ok {
						captures[run.test.Name] = obj
					}
				}
			}
		case <-done:
			return
		}
	}
}

func printer(ctx context.Context, allTests map[string][]*APIRequest, out chan *testOut, wg *sync.WaitGroup) {
	files := 0
	fails := make(map[string]struct{})
	counts := make(map[string]int)
	logs := make(map[string][]string)
	for tout := range out {
		counts[tout.file]++
		if tout.fail {
			fails[tout.file] = struct{}{}
		}
		logs[tout.file] = append(logs[tout.file], tout.logs...)
		if counts[tout.file] == len(allTests[tout.file]) {
			lines := logs[tout.file]
			delete(logs, tout.file)
			delete(counts, tout.file)
			if _, ok := fails[tout.file]; ok {
				log.Printf("--- FAIL:\t%s\n", tout.file)
			} else if tout.config.Verbose {
				log.Printf("--- PASS:\t%s\n", tout.file)
			}
			for _, line := range lines {
				log.Printf(line)
			}
			if _, ok := fails[tout.file]; !ok && tout.config.Verbose {
				log.Printf("PASS\n")
			}
			if _, ok := fails[tout.file]; ok {
				log.Printf("FAIL \n")
				log.Printf("FAIL\t%s\t\t\t%0.3fs\n", tout.file, time.Since(tout.fileStart).Seconds())
				log.Printf("FAIL \n")
			} else {
				log.Printf("ok\t%-45s\t\t%0.3fs\n", fmt.Sprintf("%s (%d tests)", tout.file, len(allTests[tout.file])), time.Since(tout.fileStart).Seconds())
			}
			files++
		}
		if files >= len(allTests) {
			wg.Done()
			return
		}
	}
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
	allTests, err := LoadTests(cfg)
	if err != nil {
		return fmt.Errorf("cannot read tests: %w", err)
	}
	if len(allTests) == 0 {
		return fmt.Errorf("no tests")
	}
	start := time.Now()
	out := make(chan *testOut)
	in := make(chan []*testIn)
	done := make(chan bool)
	var wg sync.WaitGroup
	workers := 1
	if cfg.Parallel || cfg.FileParallel {
		workers = cfg.Workers
	}
	for i := 0; i < workers; i++ {
		go worker(ctx, in, out, done)
	}
	wg.Add(1)
	go printer(ctx, allTests, out, &wg)
	for key, tests := range allTests {
		fileStart := time.Now()
		var tins []*testIn
		for _, test := range tests {
			if clients[test.Server] == nil {
				log.Fatalf("invalid server for %s ('%s')\n", test.Name, key)
				continue
			}
			tins = append(tins, &testIn{
				file:      key,
				test:      test,
				client:    clients[test.Server],
				fileStart: fileStart,
				start:     time.Now(),
				config:    cfg,
			})
			if !cfg.FileParallel {
				in <- tins
				tins = nil
			}
		}
		if cfg.FileParallel {
			in <- tins
			tins = nil
		}
	}
	wg.Wait()
	close(done)
	close(in)
	close(out)
	count := 0
	for _, value := range allTests {
		count += len(value)
	}
	log.Printf("okapi total run time: %0.3fs (%d tests total)\n", time.Since(start).Seconds(), count)
	return nil
}
