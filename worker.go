package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

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
				fmt.Printf("URL: %s status: [%s%d%s]\n", job.URL, statusColor, resp.StatusCode, ColorReset)
			case "header":
				fmt.Printf("Header: ")
				for key, values := range job.Headers {
					fmt.Printf("%s%s%s:", ColorCyan, key, ColorReset)
					for _, value := range values {
						fmt.Printf(" %s%s%s", ColorYellow, value, ColorReset)
					}
					fmt.Printf("; ")
				}
				fmt.Printf("status: [%s%d%s]\n", statusColor, resp.StatusCode, ColorReset)
			case "method":
				fmt.Printf("Method: %s %s status: [%s%d%s]\n", job.Method, job.URL, statusColor, resp.StatusCode, ColorReset)
			}

		}()
	}
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
