// Package store keeps the settings in one indented JSON file a person can read (FR-801), through the
// kit's settingsfile, which holds what every ribbon's file does alike: tolerant reading, a file that
// cannot be trusted kept aside, a file that could not be read never saved over, keys a later version
// wrote kept (NFR-C-1) and atomic writing (FR-802). This package says what WeatherRibbon's file
// holds: its keys and its cities (FR-804).
package store

import (
	"github.com/oernster/ribbonkit/infrastructure/settingsfile"
	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/settings"
)

// Store is the application's Store port over one folder.
type Store struct {
	file *settingsfile.Store[settings.Settings]
}

// New answers a store over dir, normally %APPDATA%\WeatherRibbon, for the product name, which a
// refusal to save names as the program to start again. Nothing is read or made until Load or Save
// is called.
func New(dir, name string) *Store {
	return &Store{file: settingsfile.New(dir, name, codec)}
}

// Load reads the settings (FR-801, FR-802, FR-804).
func (s *Store) Load() (application.Loaded, error) {
	loaded, notice, err := s.file.Load()
	return application.Loaded{Settings: loaded, Notice: notice}, err
}

// Save writes the settings, replacing the file whole (FR-802).
func (s *Store) Save(current settings.Settings) error { return s.file.Save(current) }
