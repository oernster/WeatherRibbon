package metno

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/oernster/weatherribbon/internal/application"
)

// fixedExpires and fixedModified are the headers the stand-in answers carry.
const (
	fixedExpires  = "Mon, 05 Oct 2026 08:00:28 GMT"
	fixedModified = "Mon, 05 Oct 2026 07:30:04 GMT"
)

// stand answers one scripted response, recording the request it was sent.
type stand struct {
	sent   *http.Request
	status int
	header http.Header
	body   []byte
	err    error
}

func (s *stand) Do(req *http.Request) (*http.Response, error) {
	s.sent = req
	if s.err != nil {
		return nil, s.err
	}
	return &http.Response{StatusCode: s.status, Header: s.header, Body: io.NopCloser(bytes.NewReader(s.body))}, nil
}

// answering answers a stand giving status with the usual headers and body.
func answering(status int, body []byte) *stand {
	header := http.Header{}
	header.Set("Expires", fixedExpires)
	header.Set("Last-Modified", fixedModified)
	return &stand{status: status, header: header, body: body}
}

// london answers the fixture's body.
func london(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile("testdata/london.json")
	if err != nil {
		t.Fatal(err)
	}
	return body
}

// fetchWith asks through client for London, answering what came back and what was logged.
func fetchWith(t *testing.T, client *stand, lastModified string) (application.Answer, string, error) {
	t.Helper()
	var log bytes.Buffer
	answer, err := NewWith("https://example.test/forecast", testUserAgent, client, &log).
		Fetch(context.Background(), application.Request{Latitude: 51.5085, Longitude: -0.1257, LastModified: lastModified})
	return answer, log.String(), err
}

// testUserAgent is the User-Agent the tests hand in; its wording is the product package's test.
const testUserAgent = "WeatherRibbon/1.2.3 https://github.com/oernster/WeatherRibbon"

// FR-301, FR-302, NFR-O-1: a request carries the User-Agent it was handed and asks for the rounded
// coordinates; each request is one line in the log.
func TestEveryRequestIdentifiesTheApplication(t *testing.T) {
	t.Parallel()
	client := answering(http.StatusOK, london(t))
	_, log, err := fetchWith(t, client, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := client.sent.Header.Get("User-Agent"); got != testUserAgent {
		t.Errorf("User-Agent %q", got)
	}
	if got := client.sent.URL.RawQuery; got != "lat=51.5085&lon=-0.1257" {
		t.Errorf("query %q", got)
	}
	if client.sent.Header.Get("If-Modified-Since") != "" {
		t.Error("a first request carried If-Modified-Since")
	}
	if log != "forecast 51.5085,-0.1257: new forecast, expires 2026-10-05T08:00:28Z\n" {
		t.Errorf("log %q", log)
	}
	if New(testUserAgent, io.Discard).baseURL != LocationforecastURL {
		t.Error("the production client does not ask Locationforecast")
	}
}

// FR-304: a request holding a cached forecast is conditional; a 304 answers NotModified with the new
// Expires.
func TestTheConditionalHeaderIsSent(t *testing.T) {
	t.Parallel()
	client := answering(http.StatusNotModified, nil)
	answer, log, err := fetchWith(t, client, fixedModified)
	if err != nil || !answer.NotModified || client.sent.Header.Get("If-Modified-Since") != fixedModified {
		t.Fatalf("answer %+v, %v, If-Modified-Since %q", answer, err, client.sent.Header.Get("If-Modified-Since"))
	}
	if log != "forecast 51.5085,-0.1257: not modified, expires 2026-10-05T08:00:28Z\n" {
		t.Errorf("log %q", log)
	}
	if want := time.Date(2026, time.October, 5, 8, 0, 28, 0, time.UTC); !answer.Expires.Equal(want) {
		t.Errorf("Expires %v; want %v", answer.Expires, want)
	}
}

// An answer is read into steps, each figure where the answer gave it; a step with no temperature is
// passed over.
func TestTheLatestForecastIsReadIntoSteps(t *testing.T) {
	t.Parallel()
	answer, _, err := fetchWith(t, answering(http.StatusOK, london(t)), "")
	if err != nil || answer.LastModified != fixedModified {
		t.Fatalf("answer %+v, %v", answer, err)
	}
	steps := answer.Forecast.Steps()
	if len(steps) != 2 {
		t.Fatalf("%d steps; want the two with a temperature", len(steps))
	}
	first := steps[0]
	if first.AirC != 14.4 || first.WindMS != 5 || first.WindFrom != 225.3 {
		t.Errorf("instant %+v", first)
	}
	if one := first.Next1; one.Hours != 1 || one.Symbol != "partlycloudy_day" || !one.HasRain || one.RainMM != 0.3 || one.HasExtremes {
		t.Errorf("1-hour block %+v", one)
	}
	if six := first.Next6; six.Hours != 6 || six.MaxC != 16.1 || six.MinC != 13 || !six.HasExtremes || six.RainMM != 1.2 {
		t.Errorf("6-hour block %+v", six)
	}
	if twelve := first.Next12; twelve.Hours != 12 || twelve.Symbol != "rain" || twelve.HasRain {
		t.Errorf("12-hour block %+v", twelve)
	}
	if second := steps[1]; second.Next1.Present() || second.WindMS != 0 || !second.Next6.HasRain || second.Next6.HasExtremes {
		t.Errorf("second step %+v", second)
	}
}

// FR-307, FR-308: a 403 is a refusal; every other unusable answer is a failure, said in the log.
func TestEveryUnusableAnswerIsAFailure(t *testing.T) {
	t.Parallel()
	if _, log, err := fetchWith(t, answering(http.StatusForbidden, nil), ""); !errors.Is(err, application.ErrRefused) || !strings.Contains(log, "refused") {
		t.Errorf("403 answered %v, logged %q", err, log)
	}
	noExpires := answering(http.StatusOK, london(t))
	noExpires.header.Del("Expires")
	cases := map[string]*stand{
		"throttled":   answering(http.StatusTooManyRequests, nil),
		"server":      answering(http.StatusInternalServerError, nil),
		"unreachable": {err: errors.New("no route to host")},
		"no expires":  noExpires,
		"too large":   answering(http.StatusOK, bytes.Repeat([]byte(" "), maxBody+1)),
		"not json":    answering(http.StatusOK, []byte("<html>")),
		"no series":   answering(http.StatusOK, []byte(`{"properties":{"timeseries":[]}}`)),
		"no props":    answering(http.StatusOK, []byte(`{}`)),
	}
	for name, client := range cases {
		answer, log, err := fetchWith(t, client, "")
		if err == nil || errors.Is(err, application.ErrRefused) || len(answer.Forecast.Steps()) != 0 || !strings.HasPrefix(log, "forecast 51.5085,-0.1257: ") || strings.HasSuffix(log, ": ok\n") {
			t.Errorf("%s: answered %v, logged %q", name, err, log)
		}
	}
}

// A body that fails part way through is a failure; a request that cannot be built is one too.
func TestABrokenAnswerIsAFailure(t *testing.T) {
	t.Parallel()
	broken := answering(http.StatusOK, nil)
	resp := &http.Response{StatusCode: http.StatusOK, Header: broken.header, Body: io.NopCloser(failingReader{})}
	if _, err := answerFrom(resp); err == nil {
		t.Error("a body failing part way through was accepted")
	}
	var log bytes.Buffer
	_, err := NewWith("://not a url", testUserAgent, broken, &log).Fetch(context.Background(), application.Request{})
	if err == nil || !strings.Contains(err.Error(), "building the request") {
		t.Errorf("a bad address answered %v", err)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("connection reset") }
