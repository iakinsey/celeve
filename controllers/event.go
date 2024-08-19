package controllers

import (
	"celeve/gateways"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

var defaultLimit = 50
var defaultOffset = 0

type getEventsParams struct {
	Limit  *int     `json:"limit"`
	Offset *int     `json:"offset"`
	Start  *int64   `json:"start"`
	End    *int64   `json:"end"`
	Tags   []string `json:"tags"`
}

type getEventParams struct {
	ID *string `json:"id"`
}

func GetEvents(eg gateways.SqliteGateway, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		log.Error().Err(err).Msg("Unable to read request body")
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	params := getEventsParams{
		Limit:  &defaultLimit,
		Offset: &defaultOffset,
	}

	if err := json.Unmarshal(body, &params); err != nil {
		log.Error().Err(err).Msg("Unable to parse request body")
		http.Error(w, "Unable to parse request body", http.StatusBadRequest)
		return
	}

	if params.Start == nil {
		start := time.Now().Unix()
		params.Start = &start
	}

	if params.End == nil {
		end := time.Now().AddDate(0, 0, 30).Unix()
		params.End = &end
	}

	events, err := eg.GetEvents(
		time.Unix(*params.Start, 0),
		time.Unix(*params.End, 0),
		*params.Limit,
		*params.Offset,
		params.Tags,
	)

	if err != nil {
		log.Error().Err(err).Msg("Unable to get events")
		http.Error(w, "Unable to get events", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON")
		w.Header().Del("Content-Type")
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func GetTags(eg gateways.SqliteGateway, w http.ResponseWriter, r *http.Request) {
	tags, err := eg.GetTags()

	if err != nil {
		log.Error().Err(err).Msg("Unable to get tags")
		http.Error(w, "Unable to get tags", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(tags); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON")
		w.Header().Del("Content-Type")
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func GetEvent(eg gateways.SqliteGateway, w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		log.Error().Err(err).Msg("Unable to read request body")
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var params getEventParams

	if err := json.Unmarshal(body, &params); err != nil {
		log.Error().Err(err).Msg("Unable to parse request body")
		http.Error(w, "Unable to parse request body", http.StatusBadRequest)
		return
	}

	if params.ID == nil {
		http.Error(w, "No event ID provided", http.StatusBadRequest)
		return
	}

	event, err := eg.GetEvent(*params.ID)

	if err != nil {
		log.Error().Err(err).Msg("Error while getting event")
		http.Error(w, "Error while getting event", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(event); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON")
		w.Header().Del("Content-Type")
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}
