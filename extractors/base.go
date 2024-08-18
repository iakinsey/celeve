package extractors

import "celeve/models"

type Extractor interface {
	GetEvent() (*models.CalendarEvent, error)
}
