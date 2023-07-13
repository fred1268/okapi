package testing

import "github.com/fred1268/go-clap/clap"

// Config holds okapi's configuration.
type Config struct {
	Servers     string `clap:"--servers-file,-s,mandatory"`
	Tests       string `clap:"trailing"`
	Timeout     int    `clap:"--timeout"`
	UserAgent   string `clap:"--user-agent"`
	ContentType string `clap:"--content-type"`
	Accept      string `clap:"--accept"`
	Verbose     bool   `clap:"--verbose,-v"`
	Parallel    bool   `clap:"--parallel,-p"`
}

// LoadConfig returns okapi's configuration from the
// command line arguments using clap, the command line
// argument parsing library.
// see: https://github.com/fred1268/go-clap
func LoadConfig(args []string) (*Config, error) {
	var cfg Config = Config{
		Timeout:     30,
		UserAgent:   "Mozilla/5.0 (compatible; okapi/1.0; +https://github.com/fred1268/okapi)",
		ContentType: "application/json",
		Accept:      "application/json",
		Parallel:    true,
	}
	if _, err := clap.Parse(args, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
