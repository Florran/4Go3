package jobs

import (
	"fmt"
	"strings"

	"github.com/florran/4go3/pkg/bypass"
	"github.com/florran/4go3/pkg/config"
)

type Job struct {
	BypassType string              //The type of bypass, (path, header or method)
	Bypass     string              //The actual bypass, (e.g., "Client-IP: 127.0.0.1" or https://example.com/../admin)
	URL        string              //URL to make the request to
	HttpMethod string              //HTTP method to use
	Headers    map[string][]string //Headers to include in the request
}

func GenerateJobs(userConfig config.Config) chan Job {
	jobs := make(chan Job, 1000)
	defaultPath := fmt.Sprintf("%s/%s", userConfig.URL, userConfig.Path)
	if userConfig.Path != "" {
		bypassPaths := bypass.GenerateBypassPaths(userConfig.URL, userConfig.Path)

		//Genrates path bypass jobs
		for _, url := range bypassPaths {
			job := Job{
				BypassType: "path",
				Bypass:     url,
				URL:        url,
				HttpMethod: "GET",
				Headers:    userConfig.Headers,
			}
			jobs <- job
		}

	}

	//Genrates header bypass jobs
	for _, header := range bypass.BypassHeaders {

		var bypassText string
		for key, values := range header {
			bypassText = fmt.Sprintf("%s: %s", key, strings.Join(values, ", "))
		}

		mergedHeaders := config.MergeHeaders(userConfig.Headers, header)
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
	for _, method := range bypass.HttpMethods {
		job := Job{
			BypassType: "method",
			Bypass:     method,
			URL:        defaultPath,
			HttpMethod: method,
			Headers:    userConfig.Headers,
		}
		jobs <- job
	}

	close(jobs)
	return jobs
}
