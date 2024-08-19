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

type lumaStrategy struct {
	config        config.LumaStrategyConfig
	userAgent     string
	channel       chan models.CalendarEvent
	tags          []string
	url           string
	pattern       *regexp.Regexp
	regionPattern *regexp.Regexp
	opts          []chromedp.ExecAllocatorOption
}

func NewLumaStrategy(params config.LumaStrategyConfig, calendarChan chan models.CalendarEvent, opts []chromedp.ExecAllocatorOption) (Job, error) {
	pattern := `^/[a-z0-9-]+$`
	regionPattern := `^/` + params.Region + `$`

	return &lumaStrategy{
		config:        params,
		userAgent:     config.Get().UserAgent,
		channel:       calendarChan,
		tags:          append(params.Tags, "luma"),
		url:           fmt.Sprintf("https://lu.ma/%s", params.Region),
		pattern:       regexp.MustCompile(pattern),
		regionPattern: regexp.MustCompile(regionPattern),
		opts:          opts,
	}, nil
}

func (s *lumaStrategy) Stop() error {
	return nil
}

func (s *lumaStrategy) Start() {
	ticker := time.NewTicker(config.Get().JobInterval)
	defer ticker.Stop()

	if err := s.perform(); err != nil {
		log.Error().Err(err).Msg("Failed luma strategy perform")
	}

	for range ticker.C {
		log.Info().Msg("Luma strategy tick")

		if err := s.perform(); err != nil {
			log.Error().Err(err).Msg("Failed luma strategy perform")
		}
	}
}

func (s *lumaStrategy) perform() error {
	log.Info().Msg("retrieving luma listing")

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(s.userAgent),
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	defer cancel()

	body, err := s.getLumaBody(ctx)

	if err != nil {
		return err
	}

	defer os.Remove(body.Name())

	urls, err := s.extractLumaUrls(body)

	if err != nil {
		return err
	}

	log.Info().Msg("Retrieving luma events")

	for _, url := range urls {
		log.Info().Msgf("Extracting url: %s", url)
		s.extractEvent(url)
	}

	return nil
}

func (s *lumaStrategy) getLumaBody(ctx context.Context) (*os.File, error) {
	var htmlContent string

	ctx, cancel := chromedp.NewContext(ctx)

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
		chromedp.Evaluate(`document.querySelector('.events').outerHTML`, &htmlContent),
	)

	if err != nil {
		return nil, err
	}

	return util.SaveHtmlBody("lumalisting", htmlContent)
}

func (s *lumaStrategy) extractLumaUrls(lumaPage *os.File) ([]string, error) {
	var result []string
	docReader := fsm.NewDocumentReaderFSM()
	docFsm := fsm.NewFSM(docReader)

	urls, err := docFsm.Perform(lumaPage)

	if err != nil {
		return nil, err
	}

	for _, u := range urls {
		if !strings.HasPrefix(u, "/") {
			continue
		}

		url, err := url.Parse(u)

		if err != nil {
			continue
		}

		if !s.pattern.MatchString(u) {
			continue
		}

		if s.regionPattern.MatchString(u) {
			continue
		}

		url.RawQuery = ""
		url.Host = "lu.ma"
		url.Scheme = "https"

		result = append(result, url.String())
	}

	return slices.Compact(result), nil
}

func (s *lumaStrategy) extractEvent(url string) {
	extractor := extractors.NewLumaExtractor(url, s.config.PrettyLocation, s.config.Tags, s.config.Timezone, s.opts)
	event, err := extractor.GetEvent()

	if err != nil {
		log.Error().Err(err).Msg("Failed to get luma event")
	} else {
		log.Info().Msgf("Pushing event %s", event.ID)
		s.channel <- *event
	}
}
