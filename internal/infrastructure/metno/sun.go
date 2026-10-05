package metno

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/forecast"
)

// SunriseURL is Sunrise 3.0's sun endpoint (section 2.3).
const SunriseURL = "https://api.met.no/weatherapi/sunrise/3.0/sun"

// sunTimeLayout is how Sunrise 3.0 writes a time: minutes and the offset, no seconds. Measured on
// 2026-10-05: London's sunrise read "2026-10-05T07:07+01:00".
const sunTimeLayout = "2006-01-02T15:04Z07:00"

// dateLayout and offsetLayout write the date and offset a request asks for.
const (
	dateLayout   = "2006-01-02"
	offsetLayout = "-07:00"
)

// noonHour is the hour of the asked-for date at which its offset is read, clear of any transition at
// night.
const noonHour = 12

// sunPayload is the part of a Sunrise 3.0 answer read. Measured on 2026-10-05: in the polar night
// (Longyearbyen, 21 December) both times are null.
type sunPayload struct {
	Properties *struct {
		Sunrise struct {
			Time *string `json:"time"`
		} `json:"sunrise"`
		Sunset struct {
			Time *string `json:"time"`
		} `json:"sunset"`
	} `json:"properties"`
}

// SunTimes is application.SunTimes over MET Norway's Sunrise 3.0.
type SunTimes struct {
	caller
	baseURL string
}

// NewSunTimes answers the production client sending userAgent, writing one line per request to log.
func NewSunTimes(userAgent string, log io.Writer) *SunTimes {
	return NewSunTimesWith(SunriseURL, userAgent, productionClient(), log)
}

// NewSunTimesWith answers a client asking baseURL through client.
func NewSunTimesWith(baseURL, userAgent string, client Doer, log io.Writer) *SunTimes {
	return &SunTimes{caller: newCaller(userAgent, client, log), baseURL: baseURL}
}

// Fetch asks for the sunrise and sunset at the coordinates on date as lived in location (FR-411).
// A time the day does not have is the zero time; a 403 is application.ErrRefused.
func (s *SunTimes) Fetch(ctx context.Context, latitude, longitude float64, date forecast.Date, location *time.Location) (application.Sun, error) {
	sun, err := s.fetch(ctx, latitude, longitude, date, location)
	s.logged(fmt.Sprintf("sun %s %s", coordinates(latitude, longitude), dateOf(date)), err)
	return sun, err
}

func (s *SunTimes) fetch(ctx context.Context, latitude, longitude float64, date forecast.Date, location *time.Location) (application.Sun, error) {
	query := at(latitude, longitude)
	query.Set("date", dateOf(date))
	noon := time.Date(date.Year, date.Month, date.Day, noonHour, 0, 0, 0, location)
	query.Set("offset", noon.Format(offsetLayout))
	resp, err := s.get(ctx, s.baseURL, query, nil, http.StatusOK)
	if err != nil {
		return application.Sun{}, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return application.Sun{}, err
	}
	var read sunPayload
	if err := json.Unmarshal(body, &read); err != nil || read.Properties == nil {
		return application.Sun{}, fmt.Errorf("the answer is not sun times: %v", err)
	}
	rise, err := sunTime(read.Properties.Sunrise.Time)
	if err != nil {
		return application.Sun{}, err
	}
	set, err := sunTime(read.Properties.Sunset.Time)
	return application.Sun{Rise: rise, Set: set}, err
}

// sunTime reads one of the answer's times; the zero time where it is null.
func sunTime(given *string) (time.Time, error) {
	if given == nil {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(sunTimeLayout, *given)
	if err != nil {
		return time.Time{}, fmt.Errorf("the time %q cannot be read: %w", *given, err)
	}
	return parsed, nil
}

// dateOf writes date as a request asks for it.
func dateOf(date forecast.Date) string {
	return time.Date(date.Year, date.Month, date.Day, 0, 0, 0, 0, time.UTC).Format(dateLayout)
}
