package main

import (
	"net/http"
)

func main() {

	config := parseFlags()

	jobs := generateJobs(config)

	client := &http.Client{
		Timeout: config.Timeout,
	}

	wg := startWorkerPool(client, jobs, config.Threads, config.Rate)

	wg.Wait()
}
