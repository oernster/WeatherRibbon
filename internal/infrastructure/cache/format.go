package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/petrichor"
)

// formatVersion is written into every entry. An entry holding any other is unreadable, so it is
// discarded and fetched again rather than misread (FR-807).
const formatVersion = 1

// Why an entry cannot be read; each is wrapped with the detail found.
var (
	errVersion  = errors.New("the entry is not in this version's format")
	errMissing  = errors.New("the entry lacks a required value")
	errExtremes = errors.New("a block gives one extreme without the other")
)

// entry is one city's file. Every optional figure is a pointer, so a figure left out is told apart
// from a zero; every required instant is a pointer, so one left out is told apart from the zero time.
type entry struct {
	Version      int        `json:"version"`
	Expires      *time.Time `json:"expires"`
	LastModified string     `json:"lastModified"`
	Fetched      *time.Time `json:"fetched"`
	Spell        spell      `json:"spell"`
	Steps        []step     `json:"steps"`
}

// spell is the city's dry spell (FR-413); Since is written only while Dry holds.
type spell struct {
	Dry   bool       `json:"dry"`
	Since *time.Time `json:"since,omitempty"`
}

// step is one forecast step. A block's span is not written: the slot it sits in says it, so the two
// can never disagree.
type step struct {
	Time     *time.Time `json:"time"`
	AirC     float64    `json:"airC"`
	WindMS   float64    `json:"windMS"`
	WindFrom float64    `json:"windFrom"`
	Next1    *block     `json:"next1,omitempty"`
	Next6    *block     `json:"next6,omitempty"`
	Next12   *block     `json:"next12,omitempty"`
}

// block is one period block of a step.
type block struct {
	Symbol string   `json:"symbol"`
	RainMM *float64 `json:"rainMM,omitempty"`
	MaxC   *float64 `json:"maxC,omitempty"`
	MinC   *float64 `json:"minC,omitempty"`
}

// encode answers cached as an entry's indented JSON.
func encode(cached application.Cached) ([]byte, error) {
	written := entry{
		Version: formatVersion, Expires: &cached.Expires, LastModified: cached.LastModified,
		Fetched: &cached.Fetched, Spell: spell{Dry: cached.Spell.Dry},
	}
	if cached.Spell.Dry {
		written.Spell.Since = &cached.Spell.Since
	}
	for _, each := range cached.Forecast.Steps() {
		written.Steps = append(written.Steps, step{
			Time: &each.Time, AirC: each.AirC, WindMS: each.WindMS, WindFrom: each.WindFrom,
			Next1: blockFrom(each.Next1), Next6: blockFrom(each.Next6), Next12: blockFrom(each.Next12),
		})
	}
	body, err := json.MarshalIndent(written, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding the forecast: %w", err)
	}
	return body, nil
}

// blockFrom answers period as written; nil where the step carries no such block.
func blockFrom(period forecast.Period) *block {
	if !period.Present() {
		return nil
	}
	written := &block{Symbol: period.Symbol}
	if period.HasRain {
		written.RainMM = &period.RainMM
	}
	if period.HasExtremes {
		written.MaxC, written.MinC = &period.MaxC, &period.MinC
	}
	return written
}

// decode reads an entry's body. Anything this version did not write is refused with the reason:
// not JSON, another format version, a required value left out or a block with half its extremes.
func decode(body []byte) (application.Cached, error) {
	var read entry
	if err := json.Unmarshal(body, &read); err != nil {
		return application.Cached{}, fmt.Errorf("the entry is not JSON of the expected shape: %w", err)
	}
	if read.Version != formatVersion {
		return application.Cached{}, fmt.Errorf("%w: version %d", errVersion, read.Version)
	}
	if read.Expires == nil || read.Fetched == nil || (read.Spell.Dry && read.Spell.Since == nil) {
		return application.Cached{}, fmt.Errorf("%w: expires, fetched or the spell's start", errMissing)
	}
	steps := make([]forecast.Step, 0, len(read.Steps))
	for _, each := range read.Steps {
		decoded, err := each.decode()
		if err != nil {
			return application.Cached{}, err
		}
		steps = append(steps, decoded)
	}
	cached := application.Cached{
		Forecast: forecast.New(steps), Expires: *read.Expires, LastModified: read.LastModified,
		Fetched: *read.Fetched, Spell: petrichor.Spell{Dry: read.Spell.Dry},
	}
	if read.Spell.Dry {
		cached.Spell.Since = *read.Spell.Since
	}
	return cached, nil
}

// decode answers the step as the domain holds it.
func (s step) decode() (forecast.Step, error) {
	if s.Time == nil {
		return forecast.Step{}, fmt.Errorf("%w: a step's time", errMissing)
	}
	decoded := forecast.Step{Time: *s.Time, AirC: s.AirC, WindMS: s.WindMS, WindFrom: s.WindFrom}
	blocks := []struct {
		read  *block
		hours int
		into  *forecast.Period
	}{
		{s.Next1, forecast.Next1Hours, &decoded.Next1},
		{s.Next6, forecast.Next6Hours, &decoded.Next6},
		{s.Next12, forecast.Next12Hours, &decoded.Next12},
	}
	for _, each := range blocks {
		period, err := each.read.decode(each.hours)
		if err != nil {
			return forecast.Step{}, err
		}
		*each.into = period
	}
	return decoded, nil
}

// decode answers the block as a period of hours; the empty Period where the step carries none.
func (b *block) decode(hours int) (forecast.Period, error) {
	if b == nil {
		return forecast.Period{}, nil
	}
	period := forecast.Period{Hours: hours, Symbol: b.Symbol}
	if b.RainMM != nil {
		period.RainMM, period.HasRain = *b.RainMM, true
	}
	if (b.MaxC == nil) != (b.MinC == nil) {
		return forecast.Period{}, errExtremes
	}
	if b.MaxC != nil {
		period.MaxC, period.MinC, period.HasExtremes = *b.MaxC, *b.MinC, true
	}
	return period, nil
}
