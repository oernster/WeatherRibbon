// Package metno asks MET Norway for forecasts (Locationforecast 2.0, CON-9) and sun times (Sunrise
// 3.0, FR-411), under its terms of service: every request names the application (FR-301) and asks
// only for coordinates rounded to four places (FR-302); a forecast request is conditional on what is
// cached (FR-304). The HTTP client is injected, so the tests never touch the network.
package metno

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/oernster/weatherribbon/internal/application"
)

// LocationforecastURL is the complete variant of Locationforecast 2.0 (section 2.3).
const LocationforecastURL = "https://api.met.no/weatherapi/locationforecast/2.0/complete"

// Forecasts is application.Forecasts over MET Norway.
type Forecasts struct {
	caller
	baseURL string
}

// New answers the production client for version, writing one line per request to log (NFR-O-1).
func New(version string, log io.Writer) *Forecasts {
	return NewWith(LocationforecastURL, version, productionClient(), log)
}

// NewWith answers a client asking baseURL through client.
func NewWith(baseURL, version string, client Doer, log io.Writer) *Forecasts {
	return &Forecasts{caller: newCaller(version, client, log), baseURL: baseURL}
}

// Fetch asks for the forecast at request's coordinates. A 304 answers NotModified with the new
// Expires; a 403 is application.ErrRefused (FR-307); any other failure is an error the service backs
// off from (FR-305, FR-308). Each request and its outcome is one line in the log.
func (f *Forecasts) Fetch(ctx context.Context, request application.Request) (application.Answer, error) {
	answer, err := f.fetch(ctx, request)
	f.logged("forecast "+coordinates(request.Latitude, request.Longitude), err)
	return answer, err
}

func (f *Forecasts) fetch(ctx context.Context, request application.Request) (application.Answer, error) {
	header := http.Header{}
	if request.LastModified != "" {
		header.Set("If-Modified-Since", request.LastModified)
	}
	resp, err := f.get(ctx, f.baseURL, at(request.Latitude, request.Longitude), header, http.StatusOK, http.StatusNotModified)
	if err != nil {
		return application.Answer{}, err
	}
	defer resp.Body.Close()
	return answerFrom(resp)
}

// answerFrom reads one accepted response into an answer.
func answerFrom(resp *http.Response) (application.Answer, error) {
	expires, err := http.ParseTime(resp.Header.Get("Expires"))
	if err != nil {
		return application.Answer{}, fmt.Errorf("no usable Expires header: %w", err)
	}
	answer := application.Answer{Expires: expires, LastModified: resp.Header.Get("Last-Modified")}
	if resp.StatusCode == http.StatusNotModified {
		answer.NotModified = true
		return answer, nil
	}
	body, err := readBody(resp)
	if err != nil {
		return application.Answer{}, err
	}
	answer.Forecast, err = decode(body)
	return answer, err
}
