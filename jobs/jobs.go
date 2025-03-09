package main

import (
	"fmt"
	"strings"
)

type Job struct {
	BypassType string              //The type of bypass, (path, header or method)
	Bypass     string              //The actual bypass, (e.g., "Client-IP: 127.0.0.1" or https://example.com/../admin)
	URL        string              //URL to make the request to
	HttpMethod string              //HTTP method to use
	Headers    map[string][]string //Headers to include in the request
}

func generateJobs(config Config) chan Job {
	jobs := make(chan Job, 1000)
	defaultPath := fmt.Sprintf("%s/%s", config.URL, config.Path)
	bypassPaths := generateBypassPaths(config.URL, config.Path)

	//Genrates header bypass jobs
	for _, url := range bypassPaths {
		job := Job{
			BypassType: "path",
			Bypass:     url,
			URL:        url,
			HttpMethod: "GET",
			Headers:    config.Headers,
		}
		jobs <- job
	}

	//Genrates header bypass jobs
	for _, header := range bypassHeaders {

		var bypassText string
		for key, values := range header {
			bypassText = fmt.Sprintf("%s: %s", key, strings.Join(values, ", "))
		}

		mergedHeaders := mergeHeaders(config.Headers, header)
		job := Job{
			BypassType: "header",
			Bypass:     bypassText,
			URL:        defaultPath,
			HttpMethod: "GET",
			Headers:    mergedHeaders,
		}
		jobs <- job
	}

	//Genrates http method bypass jobs
	for _, method := range httpMethods {
		job := Job{
			BypassType: "method",
			Bypass:     method,
			URL:        defaultPath,
			HttpMethod: method,
			Headers:    config.Headers,
		}
		jobs <- job
	}

	close(jobs)
	return jobs
}
