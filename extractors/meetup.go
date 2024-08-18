package extractors

import (
	"celeve/config"
	"celeve/models"
	"celeve/util"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/markusmobius/go-dateparser"
	"github.com/rs/zerolog/log"
)

var timeFmt = "Monday, January 2, 2006\n3:04 PM"
var timeFmt2 = "Monday, January 2, 2006 at 13:04 PM"

type meetupExtractor struct {
	url       string
	userAgent string
	location  string
	tags      []string
	tz        *time.Location
}

func NewMeetupExtractor(url string, location string, tags []string, tz string) Extractor {
	loc, err := time.LoadLocation(tz)

	if err != nil {
		log.Panic().Err(err)
	}

	return &meetupExtractor{
		url:       url,
		userAgent: config.Get().UserAgent,
		location:  location,
		tags:      append(tags, "meetup"),
		tz:        loc,
	}
}

func (s *meetupExtractor) GetEvent() (*models.CalendarEvent, error) {
	var htmlContent string
	metadata := make(map[string]string)

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(s.userAgent),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)

	defer cancel()

	var h1s []string
	var times []string

	chromedp.Run(ctx,
		chromedp.Navigate(s.url),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &htmlContent),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('h1')).map(el => el.innerText)`, &h1s),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('time')).map(el => el.innerText)`, &times),
	)

	if len(h1s) == 0 {
		return nil, fmt.Errorf("unable to extract title for url: %s", s.url)
	}

	if len(times) == 0 {
		return nil, fmt.Errorf("unable to extract time for url: %s", s.url)
	}

	metadata["raw-time"] = times[0]
	start, end, err := parseMeetupDate(times[0])

	if err != nil {
		return nil, err
	}

	event := models.CalendarEvent{
		Name:      h1s[0],
		Location:  s.location,
		StartTime: util.InjectTimezone(start, s.tz),
		EndTime:   util.InjectTimezone(end, s.tz),
		OriginURL: s.url,
		Tags:      s.tags,
		Metadata:  metadata,
	}
	event.ID = util.GetEventHash(event)

	return &event, nil
}

func parseMeetupDate(dateStr string) (time.Time, time.Time, error) {
	_, dates, err := dateparser.Search(nil, dateStr)

	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	if len(dates) == 1 {
		return dates[0].Date.Time, dates[0].Date.Time.Add(2 * time.Hour), nil
	} else if len(dates) > 1 {
		return dates[len(dates)-2].Date.Time, dates[len(dates)-1].Date.Time, nil
	}

	return time.Time{}, time.Time{}, errors.New("could not find a date")
}
