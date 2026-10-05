package structural

// NFR-S-1: the only network requests are the forecasts and sun times to api.met.no, which the metno
// package makes, plus the update check, which is ribbonkit's and held there by the same rules. Nothing else
// of WeatherRibbon's imports a network package or starts a program. The city list is built into the
// executable (DATA-1), so the search never reaches the network. The page joins these checks once it
// exists.

import (
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// forecastClient is the one folder of WeatherRibbon's own that reaches the network.
const forecastClient = "internal/infrastructure/metno"

func TestOnlyTheForecastAndUpdateImportANetworkPackage(t *testing.T) {
	structure.CheckOnlyTheExemptImportANetworkPackage(t, structure.Root(t), goFiles(t), forecastClient)
}

func TestNothingOfWeatherRibbonsStartsAProcess(t *testing.T) {
	structure.CheckOnlyNamedFilesStartAProcess(t, structure.Root(t), goFiles(t), nil)
}
