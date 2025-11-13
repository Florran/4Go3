package config

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

type Config struct {
	RawURL             string
	BaseURL            string
	PathSegments       []string
	TargetSegmentIndex int
	Threads            int
	Rate               float64
	Headers            map[string][]string
	Timeout            time.Duration
	Query              url.Values
	Fragment           string

	parsedURL url.URL
}

func flagsHelp() {
	fmt.Fprintln(os.Stderr, "Usage of 4go3:")
	fmt.Fprintln(os.Stderr, "  -url, -u string")
	fmt.Fprintln(os.Stderr, "        Target URL. Schemes will default to https when omitted (e.g. example.com)")
	fmt.Fprintln(os.Stderr, "  -segment string")
	fmt.Fprintln(os.Stderr, "        Path segment to fuzz (e.g. admin)")
	fmt.Fprintln(os.Stderr, "  -segment-index int")
	fmt.Fprintln(os.Stderr, "        Zero-based index for the segment to fuzz. Overrides -segment when supplied")
	fmt.Fprintln(os.Stderr, "  -query, -q string")
	fmt.Fprintln(os.Stderr, "        Additional query parameter in key=value form. Repeat for multiple entries")
	fmt.Fprintln(os.Stderr, "  -threads, -t int")
	fmt.Fprintln(os.Stderr, "        Number of concurrent workers (default 10)")
	fmt.Fprintln(os.Stderr, "  -rate float")
	fmt.Fprintln(os.Stderr, "        Requests per second (default 5)")
	fmt.Fprintln(os.Stderr, "  -timeout int")
	fmt.Fprintln(os.Stderr, "        Request timeout in seconds (default 10)")
	fmt.Fprintln(os.Stderr, "  -header, -H string")
	fmt.Fprintln(os.Stderr, "        Custom header in the form \"Key: Value\". Repeat for multiple entries")
}

func ParseFlags() (Config, error) {
	var (
		rawURL       string
		segmentValue string
		segmentIndex int
		timeout      int

		headers = make(map[string][]string)
		queries = url.Values{}
	)

	cfg := Config{
		Headers:            headers,
		TargetSegmentIndex: -1,
	}

	flag.Usage = flagsHelp

	flag.StringVar(&rawURL, "url", "", "Target URL (e.g. https://example.com)")
	flag.StringVar(&rawURL, "u", "", "Target URL (e.g. https://example.com)")

	flag.StringVar(&segmentValue, "segment", "", "Path segment to fuzz")
	flag.StringVar(&segmentValue, "path", "", "Alias for -segment")

	flag.IntVar(&segmentIndex, "segment-index", -1, "Zero-based index for the segment to fuzz")

	flag.IntVar(&cfg.Threads, "threads", 10, "Number of concurrent threads (workers)")
	flag.IntVar(&cfg.Threads, "t", 10, "Number of concurrent threads (workers)")

	flag.Float64Var(&cfg.Rate, "rate", 5, "Requests per second")

	flag.IntVar(&timeout, "timeout", 10, "Request timeout in seconds")
	flag.IntVar(&timeout, "max-time", 10, "Alias for timeout")

	flag.Func("query", "Query parameter (key=value)", func(q string) error {
		key, value, err := parseKeyValue(q, "=")
		if err != nil {
			return err
		}
		queries.Add(key, value)
		return nil
	})
	flag.Func("q", "Query parameter (key=value)", func(q string) error {
		key, value, err := parseKeyValue(q, "=")
		if err != nil {
			return err
		}
		queries.Add(key, value)
		return nil
	})

	flag.Func("header", "Custom header (Key: Value)", func(h string) error {
		key, value, err := parseKeyValue(h, ":")
		if err != nil {
			return err
		}
		headers[key] = append(headers[key], value)
		return nil
	})
	flag.Func("H", "Custom header (Key: Value)", func(h string) error {
		key, value, err := parseKeyValue(h, ":")
		if err != nil {
			return err
		}
		headers[key] = append(headers[key], value)
		return nil
	})

	flag.Parse()

	if rawURL == "" {
		return Config{}, errors.New("missing required -url flag")
	}

	normalizedURL := rawURL
	if !strings.Contains(normalizedURL, "://") {
		normalizedURL = "https://" + normalizedURL
	}

	parsed, err := url.Parse(normalizedURL)
	if err != nil {
		return Config{}, fmt.Errorf("invalid url: %w", err)
	}

	cfg.RawURL = normalizedURL
	cfg.BaseURL = (&url.URL{Scheme: parsed.Scheme, Host: parsed.Host}).String()
	cfg.parsedURL = *parsed

	cfg.Query = parsed.Query()
	cfg.Fragment = parsed.Fragment

	for key, values := range queries {
		for _, value := range values {
			cfg.Query.Add(key, value)
		}
	}

	cfg.parsedURL.RawQuery = ""
	cfg.parsedURL.Fragment = ""

	trimmedPath := strings.Trim(parsed.Path, "/")
	if trimmedPath != "" {
		cfg.PathSegments = strings.Split(trimmedPath, "/")
	}

	if len(cfg.PathSegments) == 0 && (segmentValue != "" || segmentIndex >= 0) {
		return Config{}, errors.New("the provided URL does not contain path segments to fuzz")
	}

	switch {
	case segmentIndex >= 0:
		if segmentIndex >= len(cfg.PathSegments) {
			return Config{}, fmt.Errorf("segment-index %d out of range", segmentIndex)
		}
		cfg.TargetSegmentIndex = segmentIndex
	case segmentValue != "":
		idx := indexOf(cfg.PathSegments, segmentValue)
		if idx == -1 {
			return Config{}, fmt.Errorf("segment %q not found in url", segmentValue)
		}
		cfg.TargetSegmentIndex = idx
	case len(cfg.PathSegments) > 0:
		cfg.TargetSegmentIndex = len(cfg.PathSegments) - 1
	}

	if cfg.Rate < 1 {
		fmt.Fprintln(os.Stderr, "rate must be greater than or equal to 1; defaulting to 1")
		cfg.Rate = 1
	}

	if cfg.Threads < 1 {
		fmt.Fprintln(os.Stderr, "threads must be greater than or equal to 1; defaulting to 1")
		cfg.Threads = 1
	}

	cfg.Timeout = time.Duration(timeout) * time.Second

	return cfg, nil
}

func (c Config) TargetSegment() string {
	if !c.HasTargetSegment() {
		return ""
	}
	return c.PathSegments[c.TargetSegmentIndex]
}

func (c Config) HasTargetSegment() bool {
	return c.TargetSegmentIndex >= 0 && c.TargetSegmentIndex < len(c.PathSegments)
}

func (c Config) PrefixSegments() []string {
	if !c.HasTargetSegment() {
		return nil
	}
	return append([]string{}, c.PathSegments[:c.TargetSegmentIndex]...)
}

func (c Config) SuffixSegments() []string {
	if !c.HasTargetSegment() {
		return nil
	}
	return append([]string{}, c.PathSegments[c.TargetSegmentIndex+1:]...)
}

func (c Config) DefaultURL() string {
	return c.buildURLWithSegments(c.PathSegments)
}

func (c Config) TargetPrefixURL() string {
	base := c.parsedURL
	base.Path = ""
	if prefix := c.PrefixSegments(); len(prefix) > 0 {
		base.Path = "/" + strings.Join(prefix, "/")
	}
	return strings.TrimSuffix(base.String(), "/")
}

func (c Config) BuildURLForVariant(variant string) string {
	if !c.HasTargetSegment() {
		return c.DefaultURL()
	}

	segments := append([]string{}, c.PathSegments...)
	segments[c.TargetSegmentIndex] = variant
	return c.buildURLWithSegments(segments)
}

func (c Config) AppendSuffixAndQuery(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	suffix := c.SuffixSegments()
	if len(suffix) > 0 {
		path := parsed.Path
		if path == "" {
			path = "/"
		}
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
		path += strings.Join(suffix, "/")
		parsed.Path = path
	}

	if len(c.Query) > 0 {
		parsed.RawQuery = encodeQuery(c.Query)
	} else {
		parsed.RawQuery = ""
	}
	parsed.Fragment = c.Fragment

	if parsed.Scheme == "" {
		parsed.Scheme = c.parsedURL.Scheme
	}
	if parsed.Host == "" {
		parsed.Host = c.parsedURL.Host
	}

	return parsed.String()
}

func (c Config) buildURLWithSegments(segments []string) string {
	builder := c.parsedURL
	if len(segments) > 0 {
		builder.Path = "/" + strings.Join(segments, "/")
	} else {
		builder.Path = ""
	}

	if len(c.Query) > 0 {
		builder.RawQuery = encodeQuery(c.Query)
	} else {
		builder.RawQuery = ""
	}
	builder.Fragment = c.Fragment

	return builder.String()
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

func parseKeyValue(input, delimiter string) (string, string, error) {
	parts := strings.SplitN(input, delimiter, 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid value %q", input)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	if key == "" || value == "" {
		return "", "", fmt.Errorf("invalid value %q", input)
	}

	return key, value, nil
}

func indexOf(values []string, target string) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

func encodeQuery(values url.Values) string {
	if values == nil {
		return ""
	}
	clone := url.Values{}
	for key, vals := range values {
		clone[key] = append([]string{}, vals...)
	}
	keys := make([]string, 0, len(clone))
	for key := range clone {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ordered := url.Values{}
	for _, key := range keys {
		for _, value := range clone[key] {
			ordered.Add(key, value)
		}
	}
	return ordered.Encode()
}
