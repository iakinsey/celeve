package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

const defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

type MeetupStrategyConfig struct {
	Query    string
	Country  string
	Province string
	City     string
	Tags     []string
	Timezone string
}

type EventbriteStrategyConfig struct {
	Query          string
	Region         string
	PrettyLocation string
	Tags           []string
}

type LumaStrategyConfig struct {
	Region         string
	Timezone       string
	PrettyLocation string
	Tags           []string
}

type ExtractorConfig struct {
	Meetup     []MeetupStrategyConfig
	Eventbrite []EventbriteStrategyConfig
	Luma       []LumaStrategyConfig
}

type Config struct {
	UserAgent          string
	EventStorePath     string
	HTTPServerAddress  string
	Extractors         ExtractorConfig
	JobInterval        time.Duration
	HTTPProxy          string
	EnableProcessorJob bool
}

func NewConfig() Config {
	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal().Err(err)
	}

	return Config{
		UserAgent:          defaultUserAgent,
		EventStorePath:     filepath.Join(cwd, "events.db"),
		HTTPServerAddress:  ":8989",
		JobInterval:        4 * time.Hour,
		EnableProcessorJob: true,
		Extractors: ExtractorConfig{
			Meetup: []MeetupStrategyConfig{
				{
					Query:    "software technology",
					Country:  "us",
					Province: "ny",
					City:     "New York",
					Tags:     []string{"tech"},
					Timezone: "America/New_York",
				},
				{
					Query:    "anime",
					Country:  "us",
					Province: "ny",
					City:     "New York",
					Tags:     []string{},
					Timezone: "America/New_York",
				},
				{
					Query:    "cosplay",
					Country:  "us",
					Province: "ny",
					City:     "New York",
					Tags:     []string{},
					Timezone: "America/New_York",
				},
				{
					Query:    "vocaloid",
					Country:  "us",
					Province: "ny",
					City:     "New York",
					Tags:     []string{},
					Timezone: "America/New_York",
				},
			},
			Eventbrite: []EventbriteStrategyConfig{
				{
					Region:         "ny--new-york",
					PrettyLocation: "New York, NY",
					Query:          "software-technology",
					Tags:           []string{"tech"},
				},
				{
					Region:         "ny--new-york",
					PrettyLocation: "New York, NY",
					Query:          "anime",
					Tags:           []string{},
				},
				{
					Region:         "ny--new-york",
					PrettyLocation: "New York, NY",
					Query:          "cosplay",
					Tags:           []string{},
				},
				{
					Region:         "ny--new-york",
					PrettyLocation: "New York, NY",
					Query:          "vocaloid",
					Tags:           []string{},
				},
			},
			Luma: []LumaStrategyConfig{
				{
					Region:         "nyc",
					Timezone:       "America/New_York",
					PrettyLocation: "New York, NY",
					Tags:           []string{},
				},
			},
		},
	}
}

var config Config = NewConfig()

func Get() Config {
	return config
}
