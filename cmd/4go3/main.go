package main

import (
	"net/http"

	"github.com/florran/4go3/internal/config"
	"github.com/florran/4go3/internal/jobs"
	"github.com/florran/4go3/internal/worker"
)

func main() {

	userConfig := config.ParseFlags()

	jobs := jobs.GenerateJobs(userConfig)

	client := &http.Client{
		Timeout: userConfig.Timeout,
	}

	wg := worker.StartWorkerPool(client, jobs, userConfig.Threads, userConfig.Rate)

	wg.Wait()
}
