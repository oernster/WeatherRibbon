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
	_ "time/tzdata"

	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/forecast"
)

// polarNight is Sunrise 3.0's answer shape for Longyearbyen on 21 December, as measured: no sunrise
// and no sunset.
const polarNight = `{"properties":{"body":"Sun","sunrise":{"time":null,"azimuth":null},"sunset":{"time":null,"azimuth":null}}}`

// october5 is the date the London answer was measured for.
var october5 = forecast.Date{Year: 2026, Month: time.October, Day: 5}

// sunWith asks through client for London's sun times on 5 October, answering what came back and what
// was logged.
func sunWith(t *testing.T, client *stand) (application.Sun, string, error) {
	t.Helper()
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}
	var log bytes.Buffer
	sun, err := NewSunTimesWith("https://example.test/sun", "1.2.3", client, &log).
		Fetch(context.Background(), 51.5085, -0.1257, october5, london)
	return sun, log.String(), err
}

// FR-411: the measured London answer reads as 07:07 and 18:29 BST; the request names the application,
// the place, the date and the date's own offset.
func TestTheSunTimesAreReadFromTheAnswer(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("testdata/sun_london.json")
	if err != nil {
		t.Fatal(err)
	}
	client := answering(http.StatusOK, body)
	sun, log, err := sunWith(t, client)
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, time.October, 5, 6, 7, 0, 0, time.UTC); !sun.Rise.Equal(want) {
		t.Errorf("sunrise %v; want %v", sun.Rise, want)
	}
	if want := time.Date(2026, time.October, 5, 17, 29, 0, 0, time.UTC); !sun.Set.Equal(want) {
		t.Errorf("sunset %v; want %v", sun.Set, want)
	}
	if got := client.sent.URL.Query(); got.Get("lat") != "51.5085" || got.Get("lon") != "-0.1257" ||
		got.Get("date") != "2026-10-05" || got.Get("offset") != "+01:00" {
		t.Errorf("query %v", got)
	}
	if client.sent.Header.Get("User-Agent") != UserAgent("1.2.3") || log != "sun 51.5085,-0.1257 2026-10-05: ok\n" {
		t.Errorf("User-Agent %q, log %q", client.sent.Header.Get("User-Agent"), log)
	}
	if NewSunTimes("1", io.Discard).baseURL != SunriseURL {
		t.Error("the production client does not ask Sunrise 3.0")
	}
}

// FR-411: a day the sun neither rises nor sets answers zero times.
func TestAPolarDayHasNoSunTimes(t *testing.T) {
	t.Parallel()
	sun, _, err := sunWith(t, answering(http.StatusOK, []byte(polarNight)))
	if err != nil || !sun.Rise.IsZero() || !sun.Set.IsZero() {
		t.Errorf("polar night %+v, %v; want no times", sun, err)
	}
}

// FR-307, FR-308: a 403 is a refusal; every other unusable answer is a failure, said in the log.
func TestEveryUnusableSunAnswerIsAFailure(t *testing.T) {
	t.Parallel()
	if _, _, err := sunWith(t, answering(http.StatusForbidden, nil)); !errors.Is(err, application.ErrRefused) {
		t.Errorf("403 answered %v", err)
	}
	for name, client := range map[string]*stand{
		"server":        answering(http.StatusServiceUnavailable, nil),
		"unreachable":   {err: errors.New("no route to host")},
		"too large":     answering(http.StatusOK, bytes.Repeat([]byte(" "), maxBody+1)),
		"not json":      answering(http.StatusOK, []byte("<html>")),
		"no properties": answering(http.StatusOK, []byte(`{}`)),
		"bad sunrise":   answering(http.StatusOK, []byte(`{"properties":{"sunrise":{"time":"soon"}}}`)),
		"bad sunset":    answering(http.StatusOK, []byte(`{"properties":{"sunrise":{"time":null},"sunset":{"time":"late"}}}`)),
	} {
		_, log, err := sunWith(t, client)
		if err == nil || strings.HasSuffix(log, ": ok\n") {
			t.Errorf("%s: answered %v, logged %q", name, err, log)
		}
	}
}
