package place

import (
	"strings"
	"time"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// Place is one place of the city list (DATA-1): a GeoNames entry with its region and country named.
type Place struct {
	// GeoNamesID is the place's id in GeoNames, what a stored city names it by (FR-801).
	GeoNamesID int
	// Name is the place's name as GeoNames writes it; ASCIIName the same in plain letters.
	Name, ASCIIName string
	// Region is the first-level region's name; empty where GeoNames names none (FR-202).
	Region string
	// Country is the country's name.
	Country string
	// Latitude and Longitude are in degrees.
	Latitude, Longitude float64
	// Population is how many live there, which orders places of equal standing (FR-202).
	Population int
	// Zone is the IANA time zone GeoNames gives the place.
	Zone string
}

// descriptionSeparator joins the parts of a place's description.
const descriptionSeparator = ", "

// Description answers how a search result names the place (FR-202): "London, England, United
// Kingdom"; "Name, Country" where GeoNames names no region.
func (p Place) Description() string {
	parts := []string{p.Name}
	if p.Region != "" {
		parts = append(parts, p.Region)
	}
	return strings.Join(append(parts, p.Country), descriptionSeparator)
}

// Label answers the label to store for what the user typed for this place (FR-204): ribbonkit's rule
// for every cell's label, trimmed and capped, with the place's name when nothing is left.
func (p Place) Label(typed string) string {
	return ribbon.Label(typed, p.Name)
}

// Clock is a city's local time as its cell shows it (FR-401).
type Clock struct {
	// Time is the local time in the chosen format, such as "08:36" or "8:36 AM".
	Time string
	// ZoneMark names the zone, as TimeRibbon's FR-203: "BST", else "UTC+5:45".
	ZoneMark string
	// OffsetSeconds is the zone's offset from UTC at the instant, what the ribbon is ordered by
	// (FR-108).
	OffsetSeconds int
}

// ClockAt answers the local time at instant in location, written in format (FR-401).
func ClockAt(instant time.Time, location *time.Location, format localtime.Format) Clock {
	local := instant.In(location)
	abbreviation, offset := local.Zone()
	return Clock{
		Time:          localtime.Text(local, format),
		ZoneMark:      localtime.ZoneMark(abbreviation, offset),
		OffsetSeconds: offset,
	}
}
