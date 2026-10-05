// Package product holds the product's name: the one home for it, read by the window, the tray, the
// settings folder, the log, the start at sign-in entry, every forecast request and the setup program.
package product

import "github.com/oernster/ribbonkit/domain/identity"

// Name is the product's name as a reader sees it.
const Name = "WeatherRibbon"

// SetupName is the setup program's name: its executable, its window class, its web view cache and
// its step log.
const SetupName = Name + "Setup"

// AppID is the reverse-domain id the Linux desktop knows WeatherRibbon by: the Flatpak's id and the
// name of its start-at-sign-in entry.
const AppID = "uk.codecrafter." + Name

// Repository is WeatherRibbon's GitHub repository as owner/name: the one the update check asks for
// its latest release (FR-603).
const Repository = "oernster/" + Name

// RepositoryURL is the repository's address, which every forecast request names as the
// application's contact; no email address is sent (FR-301, OQ-2).
const RepositoryURL = "https://github.com/" + Repository

// RibbonClass is the class the ribbon's window is created with, so it can be found by it (CON-7).
// Setup looks for it to know the ribbon is up before it closes.
const RibbonClass = Name + "Window"

// DonateURL is where the donate button at the foot of Settings sends a browser (FR-709). It is
// handed to the desktop to open rather than fetched, so it adds no request of its own.
const DonateURL = "https://www.paypal.com/ncp/payment/88LQG589TJEM6"

// App answers the names ribbonkit is handed: the one place they leave this package for the kit.
func App() identity.App { return identity.App{Name: Name, AppID: AppID} }

// Version is the version this build carries. build.ps1 stamps it from VERSION into both
// executables with -ldflags -X, which reaches only a var, never a const (CON-4). A build made any
// other way says so by carrying this placeholder.
var Version = "0.0.0-dev"

// UserAgent answers the User-Agent every request to MET Norway carries: the application, its version
// and its contact (FR-301).
func UserAgent() string { return Name + "/" + Version + " " + RepositoryURL }
