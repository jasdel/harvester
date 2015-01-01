package main

import (
	"github.com/zenazn/goji"
)

func main() {
	goji.Post("/", routeScheduleJob)
	goji.Get("/status/:jobId", routeJobStatus)
	goji.Get("/result/:jobId", routeJobResult)

	goji.Serve()
}
