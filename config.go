package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type Config struct {
	URL     string
	Path    string
	Threads int
	Rate    float64
	Headers map[string][]string
	Timeout time.Duration
}

func parseFlags() Config {
	var config Config
	var timeout int
	config.Headers = make(map[string][]string)

	flag.StringVar(&config.URL, "u", "", "Target URL (e.g., https://example.com)")

	flag.StringVar(&config.Path, "path", "", "Path to tamper (e.g., admin-panel)")

	flag.IntVar(&config.Threads, "t", 10, "Number of concurrent threads (workers) default = 10")

	flag.Float64Var(&config.Rate, "rate", 5, "Requests per second default = 5")

	flag.IntVar(&timeout, "max-time", 10, "Max time per request in seconds (default: 10)")

	flag.Func("H", "Used to set custom headers", func(h string) error {

		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid header format: %s", h)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		config.Headers[key] = append(config.Headers[key], value)

		return nil
	})

	flag.Parse()

	if config.URL == "" || config.Path == "" {
		fmt.Println("Usage: -u <URL> -path <path> [-t threads] [-rate requests per second]")
		return Config{}
	}

	if !strings.HasPrefix(config.URL, "http") {
		config.URL = "https://" + config.URL
	}

	config.URL = strings.TrimSuffix(config.URL, "/")
	config.Path = strings.Replace(config.Path, "/", "", -1)
	config.Timeout = time.Duration(timeout) * time.Second

	return config
}

func mergeHeaders(primaryHeaders, secondaryHeaders map[string][]string) map[string][]string {
	merged := make(map[string][]string)

	for key, values := range primaryHeaders {
		merged[key] = append([]string{}, values...)
	}

	for key, values := range secondaryHeaders {
		merged[key] = append([]string{}, values...)
	}

	return merged
}
