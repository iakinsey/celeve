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
)

type eventbriteExtractor struct {
	url      string
	location string
	tags     []string
	opts     []chromedp.ExecAllocatorOption
}

func NewEventbriteExtractor(url, location string, tags []string, opts []chromedp.ExecAllocatorOption) Extractor {
	return &eventbriteExtractor{
		url:      url,
		location: location,
		tags:     tags,
		opts:     opts,
	}
}

func (s *eventbriteExtractor) GetEvent() (*models.CalendarEvent, error) {
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
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`document.querySelector('h1')?.textContent || ''`, &title),
		chromedp.Evaluate(
			`document.querySelector('.summary')?.childNodes[0]?.outerHTML.trim() || ''`,
			&description,
		),
		chromedp.Evaluate(
			`Array.from(document.querySelectorAll('.date-info__full-datetime')).map(e => e.textContent.trim()).find(t => t === '' || t)`,
			&dateStr,
		),
	)

	if title == "" {
		return nil, errors.New("unable to find title")
	}

	metadata["raw-time"] = dateStr
	start, end, err := parseDateRange(dateStr)

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
		StartTime:   start,
		EndTime:     end,
		Metadata:    metadata,
	}
	event.ID = util.GetEventHash(event)

	return &event, nil
}

func parseDateRange(dateStr string) (s time.Time, e time.Time, err error) {
	defer func() {
		util.LogRecover()
	}()

	_, dates, err := dateparser.Search(nil, dateStr)

	if err != nil {
		return
	}

	if len(dates) == 1 {
		s = dates[0].Date.Time
		e = dates[0].Date.Time.Add(2 * time.Hour)
	} else if len(dates) > 1 {
		s = dates[len(dates)-2].Date.Time
		e = dates[len(dates)-1].Date.Time
	} else {
		err = errors.New("could not find a date")
	}

	return
}
