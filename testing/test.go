package testing

import (
	"context"
	"errors"
	"fmt"
	"runtime"
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

func runOne(ctx context.Context, tin *testIn, out chan<- *testOut) error {
	tout := &testOut{file: tin.file, fileStart: tin.fileStart, start: tin.start, config: tin.config}
	response, err := tin.client.Test(ctx, tin.test, tin.config.Verbose)
	if err != nil {
		if !errors.Is(err, ErrStatusCodeMismatched) && !errors.Is(err, ErrResponseMismatched) {
			tout.fail = true
			tout.logs = append(tout.logs, fmt.Sprintf("cannot run test '%s': %v", tin.file, err))
			out <- tout
			return fmt.Errorf("cannot run test '%s': %w", tin.file, err)
		}
		tout.fail = true
	}
	tout.logs = append(tout.logs, response.Logs...)
	out <- tout
	return nil
}

func worker(ctx context.Context, in chan []*testIn, out chan *testOut, done chan bool) {
	for {
		select {
		case runs := <-in:
			for _, run := range runs {
				if run.config.FileParallel {
					run.start = time.Now()
				}
				if err := runOne(ctx, run, out); err != nil {
					continue
				}
			}
		case <-done:
			return
		}
	}
}

func printer(ctx context.Context, allTests map[string][]*APIRequest, out chan *testOut, done chan bool, wg *sync.WaitGroup) {
	results := make(map[string]struct{})
	counts := make(map[string]int)
	logs := make(map[string][]string)
	for {
		select {
		case tout := <-out:
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
				} else if tout.config.Verbose {
					log.Printf("--- PASS:\t%s\n", tout.file)
				}
				for _, line := range lines {
					log.Printf(line)
				}
				if _, ok := results[tout.file]; !ok && tout.config.Verbose {
					log.Printf("PASS\n")
				}
				if _, ok := results[tout.file]; ok {
					log.Printf("FAIL \n")
					log.Printf("FAIL\t%s\t\t\t%0.3fs\n", tout.file, time.Since(tout.fileStart).Seconds())
					log.Printf("FAIL \n")
				} else {
					log.Printf("ok\t%-30s\t\t\t%0.3fs\n", tout.file, time.Since(tout.fileStart).Seconds())
				}
				if tout.config.FileParallel {
					wg.Done()
				}
			}
			if !tout.config.FileParallel {
				wg.Done()
			}
		case <-done:
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
	allTests, err := LoadTests(cfg.Tests)
	if err != nil {
		return fmt.Errorf("cannot read tests: %w", err)
	}
	start := time.Now()
	out := make(chan *testOut)
	in := make(chan []*testIn)
	done := make(chan bool)
	var wg sync.WaitGroup
	cpu := 1
	if cfg.Parallel || cfg.FileParallel {
		cpu = runtime.NumCPU()
	}
	for i := 0; i < cpu; i++ {
		go worker(ctx, in, out, done)
	}
	go printer(ctx, allTests, out, done, &wg)
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
				wg.Add(1)
				in <- tins
				tins = nil
			}
		}
		if cfg.FileParallel {
			wg.Add(1)
			in <- tins
			tins = nil
		}
	}
	wg.Wait()
	close(done)
	close(in)
	close(out)
	log.Printf("okapi total run time: %0.3fs\n", time.Since(start).Seconds())
	return nil
}
