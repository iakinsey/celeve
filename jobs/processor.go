package jobs

import (
	"celeve/config"
	"celeve/gateways"
	"celeve/models"
	"embed"
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

//go:embed keywords
var keywords embed.FS

type processorJob struct {
	tags   map[string][]string
	sqlite gateways.SqliteGateway
}

func NewProcessorJob(sqlite gateways.SqliteGateway) (Job, error) {
	tags, err := getTags()

	if err != nil {
		return nil, err
	}

	return &processorJob{
		tags:   tags,
		sqlite: sqlite,
	}, nil
}

func (s *processorJob) Start() {
	ticker := time.NewTicker(config.Get().JobInterval)
	defer ticker.Stop()

	if err := s.perform(); err != nil {
		log.Error().Err(err).Msg("Processor perform failed")
	}

	for range ticker.C {
		log.Info().Msg("Processor job tick")

		if err := s.perform(); err != nil {
			log.Error().Err(err).Msg("Processor perform failed")
		}
	}
}

func (s *processorJob) Stop() error {
	return nil
}

func (s *processorJob) perform() error {
	events, err := s.sqlite.GetEventsForProcessing()

	if len(events) == 0 {
		return nil
	}

	if err != nil {
		return err
	}

	s.hydrateTags(events)

	return s.sqlite.BulkProcessEvents(events)
}

func (s *processorJob) hydrateTags(events []*models.CalendarEvent) {
	for _, event := range events {
		tags := slices.Compact(slices.Concat(
			event.Tags,
			s.findTags(event.Description),
			s.findTags(event.Name),
		))

		if len(tags) == len(event.Tags) {
			event.Relevant = true
		}

		event.Processed = true
		event.Tags = tags
	}
}

func (s *processorJob) findTags(str string) []string {
	var matchedKeys []string

	for key, tagValues := range s.tags {
		for _, value := range tagValues {
			if strings.Contains(str, value) {
				matchedKeys = append(matchedKeys, key)
				break
			}
		}
	}
	return matchedKeys
}

func getTags() (result map[string][]string, err error) {
	var artificial_intelligence, _ = keywords.ReadFile("keywords/artificial_intelligence.json")
	var food_and_drink, _ = keywords.ReadFile("keywords/food_and_drink.json")
	var software, _ = keywords.ReadFile("keywords/software.json")
	var startup, _ = keywords.ReadFile("keywords/startup.json")
	var anime, _ = keywords.ReadFile("keywords/anime.json")
	var sports, _ = keywords.ReadFile("keywords/sports.json")
	var improv, _ = keywords.ReadFile("keywords/improv.json")
	var tags = map[string][]byte{
		"ai":           artificial_intelligence,
		"refreshments": food_and_drink,
		"software":     software,
		"startup":      startup,
		"anime":        anime,
		"sports":       sports,
		"improv":       improv,
	}
	result = make(map[string][]string)

	for k, v := range tags {
		var listing []string

		if err = json.Unmarshal(v, &listing); err != nil {
			return
		} else {
			result[k] = listing
		}
	}

	return
}
