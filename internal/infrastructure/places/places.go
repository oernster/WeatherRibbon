// Package places is the application's Places port over the bundled city list (DATA-1): the GeoNames
// extract tools/gencities writes, embedded in the executable so the search never reaches the network
// (FR-202). Zones resolve through the kit's resolver, with the tz database built in.
package places

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/oernster/ribbonkit/infrastructure/zones"
	"github.com/oernster/weatherribbon/internal/domain/place"
)

// list is the city list: a head line, then one place per line as tools/gencities writes it.
//
//go:embed cities.tsv
var list string

// commentPrefix begins the list's head line.
const commentPrefix = "#"

// Columns of each place's line.
const (
	columnID = iota
	columnName
	columnASCIIName
	columnRegion
	columnCountry
	columnLatitude
	columnLongitude
	columnPopulation
	columnZone
	columns
)

// Places is the application's Places port.
type Places struct {
	index    place.Index
	byID     map[int]place.Place
	resolver zones.Resolver
}

// New parses the embedded city list. A failure is a defect in the build rather than on the machine,
// since the list is built into the binary; a test holds it.
func New() (*Places, error) {
	return fromText(list)
}

// fromText answers the places over the list's text.
func fromText(text string) (*Places, error) {
	parsed, err := parse(text)
	if err != nil {
		return nil, err
	}
	byID := make(map[int]place.Place, len(parsed))
	for _, each := range parsed {
		byID[each.GeoNamesID] = each
	}
	return &Places{index: place.NewIndex(parsed), byID: byID}, nil
}

// Search answers the places matching typed, best first (FR-202).
func (p *Places) Search(typed string) []place.Place { return p.index.Search(typed) }

// Place answers the place with the GeoNames id; false where the list holds none (FR-803).
func (p *Places) Place(geoNamesID int) (place.Place, bool) {
	found, ok := p.byID[geoNamesID]
	return found, ok
}

// Resolve answers the location for zone; an error when the tz database does not know it.
func (p *Places) Resolve(zone string) (*time.Location, error) { return p.resolver.Resolve(zone) }

// parse reads the list in its order, passing over the head line.
func parse(text string) ([]place.Place, error) {
	var out []place.Place
	for number, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, commentPrefix) {
			continue
		}
		parsed, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("cities.tsv line %d: %w: %q", number+1, err, line)
		}
		out = append(out, parsed)
	}
	return out, nil
}

// parseLine reads one place's line.
func parseLine(line string) (place.Place, error) {
	fields := strings.Split(line, "\t")
	if len(fields) != columns {
		return place.Place{}, fmt.Errorf("it does not hold %d columns", columns)
	}
	id, errID := strconv.Atoi(fields[columnID])
	latitude, errLatitude := strconv.ParseFloat(fields[columnLatitude], 64)
	longitude, errLongitude := strconv.ParseFloat(fields[columnLongitude], 64)
	population, errPopulation := strconv.Atoi(fields[columnPopulation])
	if errID != nil || errLatitude != nil || errLongitude != nil || errPopulation != nil {
		return place.Place{}, fmt.Errorf("its id, coordinates or population are not numbers")
	}
	return place.Place{
		GeoNamesID: id, Name: fields[columnName], ASCIIName: fields[columnASCIIName],
		Region: fields[columnRegion], Country: fields[columnCountry],
		Latitude: latitude, Longitude: longitude, Population: population, Zone: fields[columnZone],
	}, nil
}
