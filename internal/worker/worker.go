package worker

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/florran/4go3/internal/jobs"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
)

func Worker(client *http.Client, jobs <-chan jobs.Job, wg *sync.WaitGroup, rateLimiter <-chan time.Time) {
	defer wg.Done()

	for job := range jobs {
		<-rateLimiter

		req, err := http.NewRequest(job.HttpMethod, job.URL, nil)
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
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Printf("[%sTIMEDOUT%s] Bypass: %s%s%s\n", ColorRed, ColorReset, ColorCyan, job.Bypass, ColorReset)
			} else {
				fmt.Printf("Response error: %v\n", err)
			}
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

			fmt.Printf("[%s%d%s] Bypass: %s%s%s\n", statusColor, resp.StatusCode, ColorReset, ColorCyan, job.Bypass, ColorReset)

		}()
	}
}

func StartWorkerPool(client *http.Client, jobs chan jobs.Job, threads int, rate float64) *sync.WaitGroup {
	var wg sync.WaitGroup

	rateLimiter := time.Tick(time.Second / time.Duration(rate))

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go Worker(client, jobs, &wg, rateLimiter)
	}

	return &wg
}
