package jobs

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/florran/4go3/internal/bypass"
	"github.com/florran/4go3/internal/config"
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
	defaultURL := userConfig.DefaultURL()

	if userConfig.HasTargetSegment() {
		for _, bypassURL := range generatePathBypassJobs(userConfig) {
			job := Job{
				BypassType: "path",
				Bypass:     bypassURL,
				URL:        bypassURL,
				HttpMethod: "GET",
				Headers:    userConfig.Headers,
			}
			jobs <- job
		}
	}

	for _, header := range bypass.HeaderPayloads {
		var bypassText string
		for key, values := range header {
			bypassText = fmt.Sprintf("%s: %s", key, strings.Join(values, ", "))
		}

		mergedHeaders := config.MergeHeaders(userConfig.Headers, header)
		job := Job{
			BypassType: "header",
			Bypass:     bypassText,
			URL:        defaultURL,
			HttpMethod: "GET",
			Headers:    mergedHeaders,
		}
		jobs <- job
	}

	for _, method := range bypass.HTTPMethods {
		job := Job{
			BypassType: "method",
			Bypass:     method,
			URL:        defaultURL,
			HttpMethod: method,
			Headers:    userConfig.Headers,
		}
		jobs <- job
	}

	close(jobs)
	return jobs
}

func generatePathBypassJobs(cfg config.Config) []string {
	target := cfg.TargetSegment()
	if target == "" {
		return nil
	}

	variants := make([]string, 0, len(bypass.PathPatterns)+1)

	variants = append(variants, cfg.BuildURLForVariant(strings.ToUpper(target)))

	prefixURL := cfg.TargetPrefixURL()
	for _, pattern := range bypass.PathPatterns {
		candidate := fmt.Sprintf(pattern, prefixURL, target)
		candidate = cfg.AppendSuffixAndQuery(candidate)
		variants = append(variants, candidate)
	}

	return deduplicate(variants)
}

func deduplicate(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var result []string
	for _, value := range values {
		normalized := normalizeURL(value)
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizeURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	parsed.Fragment = ""
	return parsed.String()
}
