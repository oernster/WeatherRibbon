package product

import (
	"strings"
	"testing"
)

// The kit is handed the product's own name and application id, the ones every other surface uses.
func TestTheKitIsHandedTheProductsNames(t *testing.T) {
	t.Parallel()
	if app := App(); app.Name != Name || app.AppID != AppID {
		t.Errorf("App answered %+v", app)
	}
}

// FR-709: the address is asserted literally, so a typo fails here rather than sending a supporter to
// a page that is not the author's.
func TestTheDonateAddressIsTheAuthorsAndSecure(t *testing.T) {
	t.Parallel()
	if DonateURL != "https://www.paypal.com/ncp/payment/88LQG589TJEM6" {
		t.Errorf("DonateURL is %q", DonateURL)
	}
	if !strings.HasPrefix(DonateURL, "https://") {
		t.Errorf("DonateURL is not https: %q", DonateURL)
	}
}

// FR-301: every request names the application, its version and its repository, asserted literally
// since MET Norway's terms ask for exactly this; no email address is sent (OQ-2).
func TestTheUserAgentNamesTheApplicationAndItsContact(t *testing.T) {
	t.Parallel()
	if got := UserAgent(); got != "WeatherRibbon/"+Version+" https://github.com/oernster/WeatherRibbon" {
		t.Errorf("UserAgent is %q", got)
	}
	if strings.Contains(UserAgent(), "@") {
		t.Errorf("UserAgent carries an address: %q", UserAgent())
	}
}
