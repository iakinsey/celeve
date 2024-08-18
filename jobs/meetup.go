package jobs

import (
	"celeve/config"
	"celeve/extractors"
	"celeve/models"
	"celeve/util"
	"celeve/util/fsm"
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

const meetupURLPrefix = "https://www.meetup.com/find/"
const urlPattern = `^\/[^\/]+\/events\/\d+\/$`

type meetupStrategy struct {
	config         config.MeetupStrategyConfig
	url            string
	userAgent      string
	channel        chan models.CalendarEvent
	prettyLocation string
	tags           []string
	opts           []chromedp.ExecAllocatorOption
}

func NewMeetupStrategy(params config.MeetupStrategyConfig, calendarChan chan models.CalendarEvent, opts []chromedp.ExecAllocatorOption) (Job, error) {
	url, err := assembleMeetupURL(params.Query, params.Country, params.Province, params.City)

	if err != nil {
		return nil, err
	}

	return &meetupStrategy{
		config:         params,
		url:            url,
		userAgent:      config.Get().UserAgent,
		channel:        calendarChan,
		prettyLocation: getPrettyLocationName(params.City, params.Province, params.Country),
		tags:           append(params.Tags, "meetup"),
		opts:           opts,
	}, nil
}

func getPrettyLocationName(country, province, city string) string {
	if province == "" {
		return fmt.Sprintf("%s, %s", city, country)
	}

	return fmt.Sprintf("%s, %s, %s", city, province, country)
}

func (s *meetupStrategy) Stop() error {
	log.Info().Msg("Stopping meetup extractor")

	return nil
}

func (s *meetupStrategy) Start() {
	ticker := time.NewTicker(config.Get().JobInterval)
	defer ticker.Stop()

	for range ticker.C {
		log.Info().Msg("Meetup strategy tick")

		if err := s.perform(); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func (s *meetupStrategy) perform() error {
	log.Info().Msg("Retrieving meetup listing")

	f, err := s.getMeetupListingBody()

	if err != nil {
		log.Info().Msg("Error was not nil")
		return err
	}

	defer os.Remove(f.Name())

	log.Info().Msg("Extracting meetup urls")

	urls, err := s.extractMeetupUrls(f)

	if err != nil {
		return err
	}

	log.Info().Msg("Retrieving meetups")

	for _, url := range urls {
		log.Info().Msgf("Extracting url: %s", url)
		s.extractEvent(url)
	}

	return nil
}

func (s *meetupStrategy) extractEvent(url string) {
	extractor := extractors.NewMeetupExtractor(url, s.prettyLocation, s.tags, s.config.Timezone)
	evt, err := extractor.GetEvent()

	if err != nil {
		log.Info().Msg(err.Error())
		return
	}

	if evt != nil {
		log.Info().Msgf("Pushing event %s", evt.ID)
		s.channel <- *evt
	} else {
		log.Info().Msg("Failed to extract event")
	}
}

func (s *meetupStrategy) extractMeetupUrls(meetupPage *os.File) ([]string, error) {
	var result []string
	docReader := fsm.NewDocumentReaderFSM()
	docFsm := fsm.NewFSM(docReader)

	urls, err := docFsm.Perform(meetupPage)
	pattern, _ := regexp.Compile(urlPattern)

	if err != nil {
		return nil, err
	}

	for _, u := range urls {
		url, err := url.Parse(u)

		if err != nil {
			continue
		}

		if url.Hostname() != "meetup.com" && url.Hostname() != "www.meetup.com" {
			continue
		}

		matched := pattern.MatchString(url.Path)

		if !matched {
			continue
		}

		url.RawQuery = ""

		result = append(result, url.String())
	}

	return slices.Compact(result), nil
}

func (s *meetupStrategy) getMeetupListingBody() (*os.File, error) {
	var htmlContent string

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), s.opts...)

	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)

	defer cancel()

	err := chromedp.Run(ctx,
		chromedp.Navigate(s.url),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &htmlContent),
	)

	if err != nil {
		return nil, err
	}

	return util.SaveHtmlBody("meetuplisting", htmlContent)
}

func assembleMeetupURL(query string, country string, province string, city string) (string, error) {
	location, err := getLocationUrlPart(country, province, city)

	if err != nil {
		return "", err
	}

	params := map[string]string{
		"suggested": "true",
		"source":    "EVENTS",
		"keywords":  url.QueryEscape(query),
		"location":  location,
		"eventType": "inPerson",
	}
	var urlQuery []string

	for key, value := range params {
		urlQuery = append(urlQuery, fmt.Sprintf("%s=%s", key, value))
	}

	return meetupURLPrefix + "?" + strings.Join(urlQuery, "&"), nil
}

func getLocationUrlPart(country string, province string, city string) (string, error) {
	code, err := util.GetISO3166Alpha2(country)

	if err != nil {
		return "", err
	}

	city = url.QueryEscape(city)

	if code == "us" {
		province = strings.ToLower(province)

		if len(province) != 2 {
			return "", fmt.Errorf("US province should be 2 characters, not %s", province)
		}

		return url.QueryEscape(fmt.Sprintf("%s--%s--%s", country, province, city)), nil
	}

	return fmt.Sprintf("%s--%s", country, city), nil
}
