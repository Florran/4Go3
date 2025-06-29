package config

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type Config struct {
	URL         string
	Path        string
	QueryParams string
	Threads     int
	Rate        float64
	Headers     map[string][]string
	Timeout     time.Duration
}

func flagsHelp() {
	fmt.Println("\n-u Target URL (e.g. https://example.com or http://example.com), if no protocol is defined (e.g. example.com) https:// is used")
	fmt.Println("\n-path Last URL segment to tamper (e.g. admin for https://example.com/panel/admin)")
	fmt.Println("\n-q Query parameters (e.g., key1=value1) use multiple flags for multiple parameters")
	fmt.Println("\n-t Number of concurrent threads (default: 10)")
	fmt.Println("\n-rate Number of recuests per second (default: 5)")
	fmt.Println("\n-max-time Max time per request in seconds")
	fmt.Println("\n-H Headers to include in requests (e.g. \"Referer: https://example.com\") use multiple flags for multiple headers")
	fmt.Println("")
}

func ParseFlags() Config {
	var userConfig Config
	var timeout int

	userConfig.Headers = make(map[string][]string)

	flag.Usage = flagsHelp

	flag.StringVar(&userConfig.URL, "u", "", "Target URL (e.g., https://example.com)")

	flag.StringVar(&userConfig.Path, "path", "", "Path to tamper (e.g., admin-panel)")

	flag.IntVar(&userConfig.Threads, "t", 10, "Number of concurrent threads (workers)")

	flag.Float64Var(&userConfig.Rate, "rate", 5, "Requests per second default ")

	flag.IntVar(&timeout, "max-time", 10, "Max time per request in seconds")

	flag.Func("q", "-q Query parameters (e.g., key1=value1) use multiple flags for multiple parameters", func(q string) error {
		parts := strings.SplitN(q, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid query parameter format: %s", q)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" || value == "" {
			return fmt.Errorf("invalid query parameter: %s", q)
		}

		if userConfig.QueryParams != "" {
			userConfig.QueryParams += "&"
		}
		userConfig.QueryParams += fmt.Sprintf("%s=%s", key, value)

		return nil
	})

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

	if userConfig.URL == "" {
		fmt.Println("Usage: -u <URL> [-path path] [-t threads] [-rate requests per second]")
		return Config{}
	}

	if !strings.HasPrefix(userConfig.URL, "http") {
		userConfig.URL = "https://" + userConfig.URL
	}

	if userConfig.Rate < 1 {
		fmt.Println("Rate must be greater than 1 defaulting to 1 (slowest)")
		userConfig.Rate = 1
	}

	if userConfig.Threads < 1 {
		fmt.Println("Threads must be greater than 1 defaulting to 1 (leasts work)")
		userConfig.Threads = 1
	}
	userConfig.URL = strings.TrimSuffix(userConfig.URL, "/")
	userConfig.Path = strings.TrimPrefix(userConfig.Path, "/")
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
