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
