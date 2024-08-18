package extractors

import (
	"celeve/models"
	"celeve/util"
	"context"
	"errors"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/chromedp/chromedp"
	"github.com/markusmobius/go-dateparser"
	"github.com/rs/zerolog/log"
)

type lumaExtractor struct {
	url      string
	location string
	tags     []string
	opts     []chromedp.ExecAllocatorOption
	tz       *time.Location
}

func NewLumaExtractor(url, location string, tags []string, tz string, opts []chromedp.ExecAllocatorOption) Extractor {
	loc, err := time.LoadLocation(tz)

	if err != nil {
		log.Panic().Err(err)
	}

	return &lumaExtractor{
		url:      url,
		location: location,
		tags:     tags,
		opts:     opts,
		tz:       loc,
	}
}

func (s *lumaExtractor) GetEvent() (*models.CalendarEvent, error) {
	var title string
	var description string
	var dateStr string
	metadata := make(map[string]string)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), s.opts...)

	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)

	defer cancel()

	chromedp.Run(ctx,
		chromedp.Navigate(s.url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(
			`Array.from(document.getElementsByClassName('title-wrapper')).map(el => el ? el.textContent.trim() : '').shift() || '';`,
			&title,
		),
		chromedp.Evaluate(
			`Array.from(document.getElementsByClassName('content')).map(e => e.outerHTML).join('\n');`,
			&description,
		),
		chromedp.Evaluate(
			`document.querySelector('.meta .title').innerText + " " + document.querySelector('.meta .desc').innerText`,
			&dateStr,
		),
	)

	if title == "" {
		return nil, errors.New("unable to find title")
	}

	if dateStr == "" {
		return nil, errors.New("unable to find date")
	}

	metadata["raw-time"] = dateStr
	start, end, err := parseLumaDate(dateStr)

	if err != nil {
		return nil, err
	}

	converter := md.NewConverter("", true, nil)
	markdown, _ := converter.ConvertString(description)
	event := models.CalendarEvent{
		Name:        title,
		Location:    s.location,
		Tags:        s.tags,
		OriginURL:   s.url,
		Description: markdown,
		StartTime:   util.InjectTimezone(start, s.tz),
		EndTime:     util.InjectTimezone(end, s.tz),
		Metadata:    metadata,
	}
	event.ID = util.GetEventHash(event)

	return &event, nil
}

func parseLumaDate(dateStr string) (time.Time, time.Time, error) {
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
