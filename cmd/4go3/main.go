package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/florran/4go3/internal/config"
	"github.com/florran/4go3/internal/jobs"
	"github.com/florran/4go3/internal/worker"
)

func main() {

	userConfig, err := config.ParseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	jobs := jobs.GenerateJobs(userConfig)

	client := &http.Client{
		Timeout: userConfig.Timeout,
	}

	wg := worker.StartWorkerPool(client, jobs, userConfig.Threads, userConfig.Rate)

	wg.Wait()
}
