// Package metno asks MET Norway's Locationforecast 2.0 for forecasts (CON-9), under its terms of
// service: every request names the application (FR-301), asks only for coordinates rounded to four
// places (FR-302) and is conditional on what is cached (FR-304). The HTTP client is injected, so the
// tests never touch the network.
package metno

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/oernster/weatherribbon/internal/application"
)

// LocationforecastURL is the complete variant of Locationforecast 2.0 (section 2.3).
const LocationforecastURL = "https://api.met.no/weatherapi/locationforecast/2.0/complete"

// Repository is the address the User-Agent names as the application's contact; no email address is
// sent (FR-301, OQ-2).
const Repository = "https://github.com/oernster/WeatherRibbon"

// requestTimeout bounds one request. A proposal: MET Norway names no figure; a request still waiting
// after this counts as failed and backs off (FR-305, FR-306).
const requestTimeout = 10 * time.Second

// maxBody is the largest answer accepted (FR-308): the measured London answer was 63,841 bytes, so
// this holds over fifteen times that. A size the server states is never trusted.
const maxBody = 1 << 20

// Doer sends one HTTP request; *http.Client is one and the tests stand in another.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Forecasts is application.Forecasts over MET Norway.
type Forecasts struct {
	baseURL   string
	userAgent string
	client    Doer
	log       io.Writer
}

// UserAgent answers the User-Agent every request carries: the application, its version and its
// contact (FR-301).
func UserAgent(version string) string {
	return "WeatherRibbon/" + version + " " + Repository
}

// New answers the production client for version, writing one line per request to log (NFR-O-1).
func New(version string, log io.Writer) *Forecasts {
	return NewWith(LocationforecastURL, version, &http.Client{Timeout: requestTimeout}, log)
}

// NewWith answers a client asking baseURL through client.
func NewWith(baseURL, version string, client Doer, log io.Writer) *Forecasts {
	return &Forecasts{baseURL: baseURL, userAgent: UserAgent(version), client: client, log: log}
}

// Fetch asks for the forecast at request's coordinates. A 304 answers NotModified with the new
// Expires; a 403 is application.ErrRefused (FR-307); any other failure is an error the service backs
// off from (FR-305, FR-308). Each request and its outcome is one line in the log.
func (f *Forecasts) Fetch(ctx context.Context, request application.Request) (application.Answer, error) {
	where := coordinates(request.Latitude, request.Longitude)
	answer, err := f.fetch(ctx, request, where)
	outcome := "ok"
	if err != nil {
		outcome = err.Error()
	}
	fmt.Fprintf(f.log, "forecast %s: %s\n", where, outcome)
	return answer, err
}

func (f *Forecasts) fetch(ctx context.Context, request application.Request, where string) (application.Answer, error) {
	query := url.Values{}
	query.Set("lat", strconv.FormatFloat(request.Latitude, 'f', -1, 64))
	query.Set("lon", strconv.FormatFloat(request.Longitude, 'f', -1, 64))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.baseURL+"?"+query.Encode(), nil)
	if err != nil {
		return application.Answer{}, fmt.Errorf("building the request for %s: %w", where, err)
	}
	req.Header.Set("User-Agent", f.userAgent)
	if request.LastModified != "" {
		req.Header.Set("If-Modified-Since", request.LastModified)
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return application.Answer{}, fmt.Errorf("asking for %s: %w", where, err)
	}
	defer resp.Body.Close()
	return answerFrom(resp)
}

// answerFrom reads one response into an answer.
func answerFrom(resp *http.Response) (application.Answer, error) {
	switch resp.StatusCode {
	case http.StatusForbidden:
		return application.Answer{}, application.ErrRefused
	case http.StatusOK, http.StatusNotModified:
	default:
		return application.Answer{}, fmt.Errorf("status %d", resp.StatusCode)
	}
	expires, err := http.ParseTime(resp.Header.Get("Expires"))
	if err != nil {
		return application.Answer{}, fmt.Errorf("no usable Expires header: %w", err)
	}
	answer := application.Answer{Expires: expires, LastModified: resp.Header.Get("Last-Modified")}
	if resp.StatusCode == http.StatusNotModified {
		answer.NotModified = true
		return answer, nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return application.Answer{}, fmt.Errorf("reading the answer: %w", err)
	}
	if len(body) > maxBody {
		return application.Answer{}, fmt.Errorf("the answer is larger than %d bytes", maxBody)
	}
	answer.Forecast, err = decode(body)
	return answer, err
}

// coordinates answers how the log names a place: its latitude and longitude as asked for.
func coordinates(latitude, longitude float64) string {
	return strconv.FormatFloat(latitude, 'f', -1, 64) + "," + strconv.FormatFloat(longitude, 'f', -1, 64)
}
