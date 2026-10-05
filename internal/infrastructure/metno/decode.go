package metno

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/oernster/weatherribbon/internal/domain/forecast"
)

// errNoTimeseries is answered for an answer holding no timeseries (FR-308).
var errNoTimeseries = errors.New("the answer holds no timeseries")

// payload is the part of a Locationforecast 2.0 answer read. Every figure is a pointer, so a figure
// the answer leaves out is told apart from a zero.
type payload struct {
	Properties *struct {
		Timeseries []stepPayload `json:"timeseries"`
	} `json:"properties"`
}

type stepPayload struct {
	Time time.Time `json:"time"`
	Data struct {
		Instant struct {
			Details struct {
				AirTemperature    *float64 `json:"air_temperature"`
				WindSpeed         *float64 `json:"wind_speed"`
				WindFromDirection *float64 `json:"wind_from_direction"`
			} `json:"details"`
		} `json:"instant"`
		Next1  *blockPayload `json:"next_1_hours"`
		Next6  *blockPayload `json:"next_6_hours"`
		Next12 *blockPayload `json:"next_12_hours"`
	} `json:"data"`
}

type blockPayload struct {
	Summary struct {
		SymbolCode string `json:"symbol_code"`
	} `json:"summary"`
	Details struct {
		PrecipitationAmount *float64 `json:"precipitation_amount"`
		AirTemperatureMax   *float64 `json:"air_temperature_max"`
		AirTemperatureMin   *float64 `json:"air_temperature_min"`
	} `json:"details"`
}

// decode reads an answer's body into a forecast. An answer that is not JSON or holds no timeseries is
// refused (FR-308); a step without an air temperature is passed over, since it cannot be shown.
func decode(body []byte) (forecast.Forecast, error) {
	var read payload
	if err := json.Unmarshal(body, &read); err != nil {
		return forecast.Forecast{}, fmt.Errorf("the answer is not a forecast: %w", err)
	}
	if read.Properties == nil || len(read.Properties.Timeseries) == 0 {
		return forecast.Forecast{}, errNoTimeseries
	}
	steps := make([]forecast.Step, 0, len(read.Properties.Timeseries))
	for _, each := range read.Properties.Timeseries {
		instant := each.Data.Instant.Details
		if instant.AirTemperature == nil {
			continue
		}
		steps = append(steps, forecast.Step{
			Time: each.Time, AirC: *instant.AirTemperature,
			WindMS: valueOf(instant.WindSpeed), WindFrom: valueOf(instant.WindFromDirection),
			Next1:  blockOf(each.Data.Next1, forecast.Next1Hours),
			Next6:  blockOf(each.Data.Next6, forecast.Next6Hours),
			Next12: blockOf(each.Data.Next12, forecast.Next12Hours),
		})
	}
	return forecast.New(steps), nil
}

// blockOf answers a block of hours from what the answer gave; the empty Period where it gave none.
func blockOf(given *blockPayload, hours int) forecast.Period {
	if given == nil {
		return forecast.Period{}
	}
	block := forecast.Period{Hours: hours, Symbol: given.Summary.SymbolCode}
	if amount := given.Details.PrecipitationAmount; amount != nil {
		block.RainMM, block.HasRain = *amount, true
	}
	if high, low := given.Details.AirTemperatureMax, given.Details.AirTemperatureMin; high != nil && low != nil {
		block.MaxC, block.MinC, block.HasExtremes = *high, *low, true
	}
	return block
}

// valueOf answers what figure points at; zero where the answer left it out.
func valueOf(figure *float64) float64 {
	if figure == nil {
		return 0
	}
	return *figure
}
