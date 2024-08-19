package gateways

import (
	"celeve/config"
	"celeve/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type SqliteGateway interface {
	UpsertEvent(models.CalendarEvent) error
	GetEvents(start, end time.Time, limit, offset int, tags []string) ([]models.CalendarEvent, error)
	GetEvent(id string) (*models.CalendarEvent, error)
	GetEventsForProcessing() ([]*models.CalendarEvent, error)
	BulkProcessEvents(events []*models.CalendarEvent) error
	GetTags() ([]string, error)
}

type sqliteGateway struct {
	db *sql.DB
}

func NewEventSqliteGateway() (SqliteGateway, error) {
	db, err := sql.Open("sqlite3", config.Get().EventStorePath)

	if err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS calendar_events (
		ID TEXT PRIMARY KEY,
		Name TEXT,
		StartTime DATETIME,
		EndTime DATETIME,
		Location TEXT,
		Description TEXT,
		OriginURL TEXT,
		Tags TEXT,
		Processed BOOLEAN,
		Relevant BOOLEAN,
		Metadata TEXT
	);
	`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	return &sqliteGateway{db: db}, nil
}

func (s *sqliteGateway) UpsertEvent(event models.CalendarEvent) error {
	query := `
	INSERT OR IGNORE INTO calendar_events (ID, Name, StartTime, EndTime, Location, Description, OriginURL, Tags, Processed, Relevant, Metadata)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	tags := strings.Trim(strings.Join(event.Tags, ","), ",")

	log.Info().Msgf("Saving event: %s", event.ID)

	meta, err := json.Marshal(event.Metadata)

	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		query,
		event.ID,
		event.Name,
		event.StartTime,
		event.EndTime,
		event.Location,
		event.Description,
		event.OriginURL,
		tags,
		false,
		false,
		string(meta),
	)

	return err
}

func (s *sqliteGateway) BulkProcessEvents(events []*models.CalendarEvent) error {
	queryTemplate := `
		UPDATE calendar_events
		SET
			Tags = CASE id
				%s
			END,
			Relevant = CASE id
				%s
			END,
			Processed = CASE id
				%s
			END
		WHERE id IN (%s)
	`
	var ids []string
	var tags []string
	var relevants []string
	var processeds []string
	caseTemplate := "WHEN '%s' THEN '%s'"
	boolCaseTemplate := "WHEN '%s' THEN %s"

	for _, event := range events {
		ids = append(
			ids,
			fmt.Sprintf("'%s'", event.ID),
		)
		tags = append(
			tags,
			fmt.Sprintf(caseTemplate, event.ID, strings.Join(event.Tags, ",")),
		)
		relevants = append(
			relevants,
			fmt.Sprintf(boolCaseTemplate, event.ID, strconv.FormatBool(event.Relevant)),
		)
		processeds = append(
			processeds,
			fmt.Sprintf(boolCaseTemplate, event.ID, strconv.FormatBool(event.Processed)),
		)
	}

	query := fmt.Sprintf(
		queryTemplate,
		strings.Join(tags, "\n"),
		strings.Join(relevants, "\n"),
		strings.Join(processeds, "\n"),
		strings.Join(ids, ", "),
	)

	_, err := s.db.Exec(query)

	return err
}

func (s *sqliteGateway) GetEvents(start, end time.Time, limit, offset int, tags []string) ([]models.CalendarEvent, error) {
	queryTemplate := `
		SELECT
			ID, Name, StartTime, EndTime, Location, Description, OriginURL, Tags, Processed, Relevant, Metadata
		FROM
			calendar_events
		WHERE StartTime BETWEEN ? AND ?
		%s
		LIMIT ? OFFSET ?;
	`
	tagsTemplate := `(Tags LIKE '%[1]s,%%' OR Tags LIKE '%%,%[1]s,%%' OR Tags LIKE '%%,%[1]s' OR Tags = '%[1]s')`
	tagBlock := ""
	var tagStatements []string

	for _, tag := range tags {
		tagStatements = append(
			tagStatements,
			fmt.Sprintf(tagsTemplate, tag),
		)
	}

	if len(tagStatements) > 0 {
		tagBlock = "AND " + strings.Join(tagStatements, " AND\n")
	}

	query := fmt.Sprintf(queryTemplate, tagBlock)

	return s.queryMany(query, start, end, limit, offset)
}

func (s *sqliteGateway) GetEventsForProcessing() ([]*models.CalendarEvent, error) {
	query := `
		SELECT
			ID, Name, StartTime, EndTime, Location, Description, OriginURL, Tags, Processed, Relevant, Metadata
		FROM
			calendar_events
		WHERE Processed = FALSE;
	`

	results, err := s.queryMany(query)

	if err != nil {
		return nil, err
	}

	var resultPtrs []*models.CalendarEvent

	for _, event := range results {
		resultPtrs = append(resultPtrs, &event)
	}

	return resultPtrs, nil
}

func (s *sqliteGateway) GetEvent(id string) (*models.CalendarEvent, error) {
	query := `
		SELECT
			ID, Name, StartTime, EndTime, Location, Description, OriginURL, Tags, Processed, Relevant, Metadata
		FROM
			calendar_events
		WHERE ID = ?
		LIMIT 1
	`
	var event models.CalendarEvent
	var tags string
	var rawMeta string

	row := s.db.QueryRow(query, id)
	err := row.Scan(
		&event.ID,
		&event.Name,
		&event.StartTime,
		&event.EndTime,
		&event.Location,
		&event.Description,
		&event.OriginURL,
		&tags,
		&event.Processed,
		&event.Relevant,
		&rawMeta,
	)

	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(rawMeta), &event.Metadata); err != nil {
		return nil, err
	}

	event.Tags = strings.Split(strings.Trim(tags, ","), ",")

	return &event, nil
}

func (s *sqliteGateway) GetTags() ([]string, error) {
	query := `
		WITH RECURSIVE split_string AS (
			SELECT 
				substr(Tags, 1, instr(Tags || ',', ',') - 1) AS part,
				substr(Tags, instr(Tags || ',', ',') + 1) AS rest,
				StartTime
			FROM calendar_events
			WHERE StartTime > ?
			UNION ALL
			SELECT 
				substr(rest, 1, instr(rest || ',', ',') - 1),
				substr(rest, instr(rest || ',', ',') + 1),
				StartTime
			FROM split_string
			WHERE rest != '' AND StartTime > ?
		),
		deduplicated_tags AS (
			SELECT DISTINCT part FROM split_string
		)
		SELECT part
		FROM deduplicated_tags
		WHERE part != "";
	`

	t := time.Now()
	rows, err := s.db.Query(query, t, t)

	if err != nil {
		return nil, err
	}

	var tags []string

	for rows.Next() {
		var tag string

		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *sqliteGateway) queryMany(query string, args ...any) ([]models.CalendarEvent, error) {
	rows, err := s.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	var events []models.CalendarEvent

	for rows.Next() {
		var event models.CalendarEvent
		var tags string
		var rawMeta string

		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.StartTime,
			&event.EndTime,
			&event.Location,
			&event.Description,
			&event.OriginURL,
			&tags,
			&event.Processed,
			&event.Relevant,
			&rawMeta,
		)

		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal([]byte(rawMeta), &event.Metadata); err != nil {
			return nil, err
		}

		if tags != "" {
			event.Tags = strings.Split(strings.Trim(tags, ","), ",")
		} else {
			event.Tags = make([]string, 0)
		}

		events = append(events, event)
	}

	return events, nil
}
