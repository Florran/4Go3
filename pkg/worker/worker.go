package worker

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/florran/4go3/pkg/jobs"
	"github.com/florran/4go3/pkg/utils"
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
				fmt.Printf("Bypass: %s%s%s status: [%sTIMEDOUT%s]\n", utils.ColorCyan, job.Bypass, utils.ColorReset, utils.ColorRed, utils.ColorReset)
			} else {
				fmt.Printf("Response error: %v\n", err)
			}
			continue
		}

		func() {
			defer resp.Body.Close()
			statusColor := utils.ColorYellow

			switch resp.StatusCode {
			case http.StatusOK:
				statusColor = utils.ColorGreen

			case http.StatusForbidden:
				statusColor = utils.ColorRed

			case http.StatusNotFound:
				statusColor = utils.ColorYellow

			}

			fmt.Printf("Bypass: %s%s%s status: [%s%d%s]\n", utils.ColorCyan, job.Bypass, utils.ColorReset, statusColor, resp.StatusCode, utils.ColorReset)

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
