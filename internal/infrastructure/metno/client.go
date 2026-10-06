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

// requestTimeout bounds one request. A proposal: MET Norway names no figure; a request still waiting
// after this counts as failed and backs off (FR-305, FR-306).
const requestTimeout = 10 * time.Second

// maxBody is the largest answer accepted (FR-308): the measured London forecast was 63,841 bytes, so
// this holds over fifteen times that. A size the server states is never trusted.
const maxBody = 1 << 20

// Doer sends one HTTP request; *http.Client is one and the tests stand in another.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// caller is what every request to MET Norway shares: who is asking, through what, logged where. The
// User-Agent naming the application is the product's to word (FR-301); every request carries it.
type caller struct {
	userAgent string
	client    Doer
	log       io.Writer
}

// newCaller answers a caller sending userAgent through client, logging to log.
func newCaller(userAgent string, client Doer, log io.Writer) caller {
	return caller{userAgent: userAgent, client: client, log: log}
}

// productionClient answers the HTTP client every production request goes through.
func productionClient() Doer { return &http.Client{Timeout: requestTimeout} }

// get sends a GET for address with query and header, naming the application (FR-301). A 403 is
// application.ErrRefused (FR-307); any status but those in accepted is an error. The caller closes
// the body of the response answered.
func (c caller) get(ctx context.Context, address string, query url.Values, header http.Header, accepted ...int) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("building the request: %w", err)
	}
	for name, values := range header {
		req.Header[name] = values
	}
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("asking: %w", err)
	}
	if resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		return nil, application.ErrRefused
	}
	for _, each := range accepted {
		if resp.StatusCode == each {
			return resp, nil
		}
	}
	resp.Body.Close()
	return nil, fmt.Errorf("status %d", resp.StatusCode)
}

// logged writes one line naming what was asked and how it went (NFR-O-1).
func (c caller) logged(what string, err error) {
	outcome := "ok"
	if err != nil {
		outcome = err.Error()
	}
	c.wrote(what, outcome)
}

// wrote writes one line naming what was asked and the outcome given.
func (c caller) wrote(what, outcome string) {
	fmt.Fprintf(c.log, "%s: %s\n", what, outcome)
}

// readBody reads resp's body, refusing one larger than maxBody rather than cutting it short (FR-308).
func readBody(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, fmt.Errorf("reading the answer: %w", err)
	}
	if len(body) > maxBody {
		return nil, fmt.Errorf("the answer is larger than %d bytes", maxBody)
	}
	return body, nil
}

// at answers the query naming a place by latitude and longitude, written as asked for (FR-302).
func at(latitude, longitude float64) url.Values {
	query := url.Values{}
	query.Set("lat", degrees(latitude))
	query.Set("lon", degrees(longitude))
	return query
}

// coordinates answers how the log names a place: its latitude and longitude as asked for.
func coordinates(latitude, longitude float64) string {
	return degrees(latitude) + "," + degrees(longitude)
}

// degrees writes a coordinate in its shortest form, already rounded by the application.
func degrees(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }
