// Package cache keeps each city's last forecast on disk, one file per city, so an offline start
// shows the last forecasts with their age and the dry spells survive a restart (FR-807, FR-413).
//
// A file is named after its city id, hex encoded, so any id a hand-edited settings file holds makes
// a safe name and two ids differing only in letter case never share a file. Every write replaces a
// file whole through atomicfile. An entry that cannot be read is discarded, so it is fetched again.
package cache

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/oernster/ribbonkit/infrastructure/atomicfile"
	"github.com/oernster/weatherribbon/internal/application"
)

// Folder is the cache's folder inside the settings folder.
const Folder = "forecasts"

// extension ends every entry's file name.
const extension = ".json"

// MaxIDLength answers the longest city id kept, in bytes: the longest whose hex-encoded name
// atomicfile can write.
func MaxIDLength() int { return (atomicfile.MaxNameLength() - len(extension)) / 2 }

// maxEntry is the largest entry read. Measured 2026-10-05: a full London answer of 88 steps, 38,562
// bytes from MET Norway, made a 25,926-byte entry, so this holds about forty times that. A file's
// stated size is never trusted.
const maxEntry = 1 << 20

// Permissions for what the cache creates: the folder and its files are the user's alone.
const (
	folderMode fs.FileMode = 0o700
	fileMode   fs.FileMode = 0o600
)

// ErrNoID is answered for an empty city id, which names no file.
var ErrNoID = errors.New("the city id is empty")

// ErrIDTooLong is answered for a city id longer than MaxIDLength.
var ErrIDTooLong = fmt.Errorf("the city id is longer than %d bytes", MaxIDLength())

// errTooLarge is why an entry over maxEntry is discarded.
var errTooLarge = fmt.Errorf("the entry is larger than %d bytes", maxEntry)

// Cache is the application's Cache port over one folder.
type Cache struct {
	dir string
	log io.Writer
}

// New answers a cache in settingsDir's Folder, logging each discarded entry to log. Nothing is read
// or made until it is asked.
func New(settingsDir string, log io.Writer) *Cache {
	return &Cache{dir: filepath.Join(settingsDir, Folder), log: log}
}

// Load answers every readable entry by city id (FR-807). No folder means no entries. An entry that
// cannot be read is discarded, logged and left out, so it is fetched again; a file whose name no id
// gives is not the cache's and is left alone. An error is a fault listing the folder.
func (c *Cache) Load() (map[string]application.Cached, error) {
	found := map[string]application.Cached{}
	listed, err := os.ReadDir(c.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return found, nil
	}
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", c.dir, err)
	}
	for _, each := range listed {
		id, ok := idOf(each)
		if !ok {
			continue
		}
		cached, err := read(filepath.Join(c.dir, each.Name()))
		if err != nil {
			c.discard(id, each.Name(), err)
			continue
		}
		found[id] = cached
	}
	return found, nil
}

// idOf answers the city id an entry's file name gives; false where the name is not one the cache
// writes.
func idOf(listed fs.DirEntry) (string, bool) {
	stem, ok := strings.CutSuffix(listed.Name(), extension)
	if !ok || !listed.Type().IsRegular() {
		return "", false
	}
	id, err := hex.DecodeString(stem)
	if err != nil || len(id) == 0 || nameOf(string(id)) != listed.Name() {
		return "", false
	}
	return string(id), true
}

// read reads and decodes one entry, refusing one larger than maxEntry rather than reading it all.
func read(path string) (application.Cached, error) {
	body, err := readCapped(path)
	if err != nil {
		return application.Cached{}, err
	}
	if len(body) > maxEntry {
		return application.Cached{}, errTooLarge
	}
	return decode(body)
}

// readCapped answers at most one byte more than maxEntry of the file at path, which is enough to
// tell an entry too large.
func readCapped(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(io.LimitReader(file, maxEntry+1))
}

// discard removes an entry that could not be read and logs why, with whether the removal failed.
func (c *Cache) discard(id, name string, reason error) {
	line := fmt.Sprintf("cache %s: discarded: %v", id, reason)
	if err := os.Remove(filepath.Join(c.dir, name)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		line += fmt.Sprintf("; not removed: %v", err)
	}
	fmt.Fprintln(c.log, line)
}

// Save writes one city's entry atomically, making the folder where it is not there (FR-807).
func (c *Cache) Save(cityID string, cached application.Cached) error {
	path, err := c.path(cityID)
	if err != nil {
		return err
	}
	body, err := encode(cached)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.dir, folderMode); err != nil {
		return fmt.Errorf("making %s: %w", c.dir, err)
	}
	return atomicfile.Write(path, body, fileMode)
}

// Forget removes one city's entry; an entry that is not there is already forgotten (FR-205, FR-207).
func (c *Cache) Forget(cityID string) error {
	path, err := c.path(cityID)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("forgetting %s: %w", cityID, err)
	}
	return nil
}

// path answers the entry's path for cityID: ErrNoID for an empty id, ErrIDTooLong where its name
// would not fit.
func (c *Cache) path(cityID string) (string, error) {
	if cityID == "" {
		return "", ErrNoID
	}
	if len(cityID) > MaxIDLength() {
		return "", ErrIDTooLong
	}
	return filepath.Join(c.dir, nameOf(cityID)), nil
}

// nameOf answers the file name of cityID's entry.
func nameOf(cityID string) string { return hex.EncodeToString([]byte(cityID)) + extension }
