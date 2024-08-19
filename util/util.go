package util

import (
	"celeve/models"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/biter777/countries"
	"github.com/rs/zerolog/log"
)

func GetISO3166Alpha2(keyword string) (string, error) {
	code := countries.ByName(keyword)

	if code == countries.Unknown {
		return "", fmt.Errorf("Unknown country %s", keyword)
	}

	return strings.ToLower(code.Alpha2()), nil
}

func GetEventHash(event models.CalendarEvent) string {
	sort.Strings(event.Tags)

	data := fmt.Sprintf("%s%s%s%s%s%s",
		event.Name,
		event.StartTime.Format(time.RFC3339),
		event.EndTime.Format(time.RFC3339),
		event.Location,
		event.Description,
		event.OriginURL,
	)

	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}

func SaveHtmlBody(name, htmlContent string) (*os.File, error) {
	log.Info().Msg("Saving body")

	pid := os.Getpid()
	t := time.Now().Unix()
	fname := fmt.Sprintf("%s-%d-%d", name, pid, t)
	tmpFile, err := ioutil.TempFile(os.TempDir(), fname)

	if err != nil {
		return nil, err
	}

	if _, err = tmpFile.WriteString(htmlContent); err != nil {
		return nil, err
	}

	tmpFile.Seek(0, 0)
	log.Info().Msg("Meetup listing body saved")

	return tmpFile, nil
}

func InjectTimezone(t time.Time, tz *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), tz)
}

func RecoverError() error {
	r := recover()

	switch x := r.(type) {
	case nil:
		return nil
	case string:
		return errors.New(x)
	case error:
		return x
	default:
		return fmt.Errorf("unknown panic: %v", r)
	}
}

func LogRecover() {
	if err := RecoverError(); err != nil {
		log.Error().Err(err).Msg("Panic recovered")
	}
}
