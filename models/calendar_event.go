package models

import "time"

type CalendarEvent struct {
	ID          string
	Name        string
	StartTime   time.Time
	EndTime     time.Time
	Location    string
	Description string
	OriginURL   string
	Tags        []string
	Processed   bool
	Relevant    bool
	Metadata    map[string]string
}
