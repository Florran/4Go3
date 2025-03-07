package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
)

func main() {

	domain := flag.String("d", "", "Target URL (e.g., google.com or https://google.com)")

	var headers []string
	flag.Func("H", "Used to set custom headers", func(h string) error {
		if !strings.Contains(h, ":") {
			return fmt.Errorf("invalid header format: %s", h)
		}
		headers = append(headers, h)
		return nil
	})

	flag.Parse()

	if *domain == "" {
		fmt.Printf("-d flag must be provided\n Example: -d example.com")
		return
	}

	if !strings.HasPrefix(*domain, "http") && !strings.HasPrefix(*domain, "https") {
		*domain = "http://" + *domain
	}

	client := http.Client{}
	req, err := http.NewRequest("GET", *domain, nil)
	if err != nil {
		fmt.Printf("Error creating request:\n %v", err)
		return
	}

	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			req.Header.Add(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request:\n %v", err)
		return
	}

	for key, value := range req.Header {
		for _, i := range value {
			fmt.Printf("%s: %s\n", key, i)
		}
	}

	for key, value := range resp.Header {
		for _, i := range value {
			fmt.Printf("%s: %s\n", key, i)
		}
	}
}
