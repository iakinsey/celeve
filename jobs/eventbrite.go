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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
)

var processUrlBatchSize = 5
var extractEventBatchSize = 25
var ebUrlRegex = regexp.MustCompile(`https?:\/\/(www\.)?eventbrite\.[^\/]+\/e\/.*`)

const eventbriteUrlBase = "https://www.eventbrite.com/d/%s/%s/?page=%d"
const getEventbritePages = `(() => {
	const element = document.querySelector('[data-testid="pagination-parent"]');
	const span = element.querySelector('span');
	if (span) {
		span.remove();
	}
	const text = element.textContent.trim();
	const number = text.match(/\d+/)[0];
	return number;
})()`

type eventbriteStrategy struct {
	config    config.EventbriteStrategyConfig
	userAgent string
	channel   chan models.CalendarEvent
	tags      []string
	opts      []chromedp.ExecAllocatorOption
}

func NewEventbriteStrategy(params config.EventbriteStrategyConfig, calendarChan chan models.CalendarEvent, opts []chromedp.ExecAllocatorOption) (Job, error) {
	return &eventbriteStrategy{
		config:    params,
		userAgent: config.Get().UserAgent,
		channel:   calendarChan,
		tags:      append(params.Tags, "eventbrite"),
		opts:      opts,
	}, nil
}

func (s *eventbriteStrategy) Start() {
	ticker := time.NewTicker(config.Get().JobInterval)
	defer ticker.Stop()

	for range ticker.C {
		log.Info().Msg("Eventbrite strategy tick")

		if err := s.perform(); err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func (s *eventbriteStrategy) perform() error {
	log.Info().Msg("Retrieving eventbrite listing")

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(s.userAgent),
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	defer cancel()

	body, pages, err := s.getEventbriteBody(ctx, 1)

	if err != nil && body == nil {
		return err
	} else if err != nil {
		log.Error().Err(err)
	}

	log.Info().Msg("Retrieving eventbrite events")

	var wg sync.WaitGroup
	sem := make(chan bool, processUrlBatchSize)
	c := make(chan string)

	go func() {
		wg.Wait()
		close(c)
	}()

	go func() {
		for i := 2; i <= pages; i++ {
			sem <- true
			wg.Add(1)
			go s.processUrl(ctx, i, &wg, c, sem)
		}
	}()

	var urls []string

	for item := range c {
		urls = append(urls, item)

		if len(urls) == extractEventBatchSize {
			s.extractEventbriteEvents(urls)
			urls = nil
		}
	}

	if len(urls) > 0 {
		s.extractEventbriteEvents(urls)
	}

	return nil
}

func (s *eventbriteStrategy) processUrl(ctx context.Context, i int, wg *sync.WaitGroup, c chan string, sem chan bool) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Error().Any("panic", r)
		}
		<-sem
	}()

	body, _, err := s.getEventbriteBody(ctx, i)

	if err != nil {
		log.Error().Err(err)
		return
	}

	urls, err := s.extractEventbriteUrls(body)

	if osErr := os.Remove(body.Name()); osErr != nil {
		log.Error().Err(err)
	}

	if err != nil {
		log.Error().Err(err)
		return
	}

	for _, url := range urls {
		c <- url
	}
}

func (s *eventbriteStrategy) Stop() error {
	return nil
}

func (s *eventbriteStrategy) assembleEventbriteURL(page int) string {
	return fmt.Sprintf(eventbriteUrlBase, s.config.Region, s.config.Query, page)
}

func (s *eventbriteStrategy) extractEventbriteEvents(urls []string) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Any("panic", r)
		}
	}()

	for _, url := range urls {
		extractor := extractors.NewEventbriteExtractor(url, s.config.PrettyLocation, s.config.Tags, s.opts)
		event, err := extractor.GetEvent()

		if err != nil {
			log.Error().Err(err)
		} else {
			log.Info().Msgf("Pushing event %s", event.ID)
			s.channel <- *event
		}
	}
}

func (s *eventbriteStrategy) extractEventbriteUrls(body *os.File) ([]string, error) {
	var result []string
	docReader := fsm.NewDocumentReaderFSM()
	docFsm := fsm.NewFSM(docReader)

	urls, err := docFsm.Perform(body)

	if err != nil {
		return nil, err
	}

	for _, u := range urls {
		url, err := url.Parse(u)

		if err != nil {
			continue
		}

		if !strings.Contains(url.Hostname(), "eventbrite") {
			continue
		}

		matched := ebUrlRegex.MatchString(u)

		if !matched {
			continue
		}

		url.RawQuery = ""

		result = append(result, url.String())
	}

	return slices.Compact(result), nil
}

func (s *eventbriteStrategy) getEventbriteBody(ctx context.Context, page int) (*os.File, int, error) {
	var htmlContent string
	url := s.assembleEventbriteURL(page)
	ctx, cancel := chromedp.NewContext(ctx)

	defer cancel()

	var pagesStr string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.WaitReady("body"),
		chromedp.OuterHTML("html", &htmlContent),
		chromedp.Evaluate(getEventbritePages, &pagesStr),
	)

	if err != nil {
		return nil, 0, err
	}

	pages, err := strconv.Atoi(pagesStr)

	if err != nil {
		log.Error().Err(err)
	}

	body, err := util.SaveHtmlBody("eventbritelisting", htmlContent)

	return body, pages, err
}
