package main

import "fmt"

type Job struct {
	BypassType string
	URL        string
	Method     string
	Headers    map[string][]string
}

func generateJobs(config Config) chan Job {
	jobs := make(chan Job, 1000)
	defaultPath := fmt.Sprintf("%s/%s", config.URL, config.Path)

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

	for _, header := range bypassHeaders {
		mergedHeaders := mergeHeaders(config.Headers, header)
		job := Job{
			BypassType: "header",
			URL:        defaultPath,
			Method:     "GET",
			Headers:    mergedHeaders,
		}
		jobs <- job
	}

	for _, method := range httpMethods {
		job := Job{
			BypassType: "method",
			URL:        defaultPath,
			Method:     method,
			Headers:    config.Headers,
		}
		jobs <- job
	}
	close(jobs)
	return jobs
}
