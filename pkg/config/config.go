package config

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

func ParseFlags() Config {
	var userConfig Config
	var timeout int
	userConfig.Headers = make(map[string][]string)

	flag.StringVar(&userConfig.URL, "u", "", "Target URL (e.g., https://example.com)")

	flag.StringVar(&userConfig.Path, "path", "", "Path to tamper (e.g., admin-panel)")

	flag.IntVar(&userConfig.Threads, "t", 10, "Number of concurrent threads (workers)")

	flag.Float64Var(&userConfig.Rate, "rate", 5, "Requests per second default ")

	flag.IntVar(&timeout, "max-time", 10, "Max time per request in seconds")

	flag.Func("H", "Used to set custom headers", func(h string) error {

		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid header format: %s", h)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		userConfig.Headers[key] = append(userConfig.Headers[key], value)

		return nil
	})

	flag.Parse()

	if userConfig.URL == "" || userConfig.Path == "" {
		fmt.Println("Usage: -u <URL> -path <path> [-t threads] [-rate requests per second]")
		return Config{}
	}

	if !strings.HasPrefix(userConfig.URL, "http") {
		userConfig.URL = "https://" + userConfig.URL
	}

	userConfig.URL = strings.TrimSuffix(userConfig.URL, "/")
	userConfig.Path = strings.Replace(userConfig.Path, "/", "", -1)
	userConfig.Timeout = time.Duration(timeout) * time.Second

	return userConfig
}

func MergeHeaders(primaryHeaders, secondaryHeaders map[string][]string) map[string][]string {
	merged := make(map[string][]string)

	for key, values := range primaryHeaders {
		merged[key] = append([]string{}, values...)
	}

	for key, values := range secondaryHeaders {
		merged[key] = append([]string{}, values...)
	}

	return merged
}
