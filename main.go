package main

import (
	"celeve/config"
	"celeve/controllers"
	"celeve/gateways"
	"celeve/jobs"
	"celeve/models"
	"celeve/util"
	"net/http"

	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func startJobServer(gateway gateways.SqliteGateway) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	conf := config.NewConfig()
	calendarChan := make(chan models.CalendarEvent)
	var jobsToRun []jobs.Job
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(config.Get().UserAgent),
	)

	if config.Get().HTTPProxy != "" {
		opts = append(
			opts,
			chromedp.ProxyServer(config.Get().HTTPProxy),
		)
	}

	/////////////////////////////////////////////////////////////////////////
	// Meetup
	/////////////////////////////////////////////////////////////////////////

	for _, meetupConfig := range conf.Extractors.Meetup {
		job, err := jobs.NewMeetupStrategy(meetupConfig, calendarChan, opts)

		if err != nil {
			log.Fatal().Err(err)
		}

		jobsToRun = append(jobsToRun, job)
	}

	/////////////////////////////////////////////////////////////////////////
	// Eventbrite
	/////////////////////////////////////////////////////////////////////////

	for _, ebConfig := range conf.Extractors.Eventbrite {
		job, err := jobs.NewEventbriteStrategy(ebConfig, calendarChan, opts)

		if err != nil {
			log.Fatal().Err(err)
		}

		jobsToRun = append(jobsToRun, job)
	}
	/////////////////////////////////////////////////////////////////////////
	// Luma
	/////////////////////////////////////////////////////////////////////////

	for _, lumaConfig := range conf.Extractors.Luma {
		job, err := jobs.NewLumaStrategy(lumaConfig, calendarChan, opts)

		if err != nil {
			log.Fatal().Err(err)
		}

		jobsToRun = append(jobsToRun, job)
	}

	/////////////////////////////////////////////////////////////////////////
	// Processor
	/////////////////////////////////////////////////////////////////////////

	if config.Get().EnableProcessorJob {
		job, err := jobs.NewProcessorJob(gateway)

		if err != nil {
			log.Fatal().Err(err)
		}

		jobsToRun = append(jobsToRun, job)
	}

	/////////////////////////////////////////////////////////////////////////
	// Start Jobs
	/////////////////////////////////////////////////////////////////////////

	for _, job := range jobsToRun {
		go executeJob(job.Start)
	}

	/////////////////////////////////////////////////////////////////////////
	// Process Events
	/////////////////////////////////////////////////////////////////////////

	for event := range calendarChan {
		gateway.UpsertEvent(event)
	}
}

func executeJob(fn func()) {
	for {
		func() {
			defer func() {
				util.LogRecover()
			}()
			fn()
		}()
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func startHttpServer(gateway gateways.SqliteGateway) {
	mux := http.NewServeMux()

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		controllers.GetEvents(gateway, w, r)
	})
	mux.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		controllers.GetEvent(gateway, w, r)
	})
	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		controllers.GetTags(gateway, w, r)
	})

	handler := enableCORS(mux)

	log.Info().Msgf("HTTP server listening on %s", config.Get().HTTPServerAddress)
	http.ListenAndServe(config.Get().HTTPServerAddress, handler)
}

func main() {
	gateway, err := gateways.NewEventSqliteGateway()

	if err != nil {
		log.Fatal().Err(err)
	}

	go startJobServer(gateway)
	startHttpServer(gateway)
}
