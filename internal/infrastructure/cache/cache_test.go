package cache

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// A city id as the kit's IDs answers them.
const (
	london = "Q2ZJ7L4MVXH3KSD5TW6RYBNPEA"
	oslo   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// newCache answers a cache in a fresh settings folder with the log it writes to.
func newCache(t *testing.T) (*Cache, *bytes.Buffer) {
	t.Helper()
	log := &bytes.Buffer{}
	return New(t.TempDir(), log), log
}

// put writes body as id's entry, making the folder.
func put(t *testing.T, c *Cache, id, body string) {
	t.Helper()
	if err := os.MkdirAll(c.dir, folderMode); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(c.dir, nameOf(id)), []byte(body), fileMode); err != nil {
		t.Fatal(err)
	}
}

// listing answers the names in the cache's folder.
func listing(t *testing.T, c *Cache) []string {
	t.Helper()
	listed, err := os.ReadDir(c.dir)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(listed))
	for _, each := range listed {
		names = append(names, each.Name())
	}
	return names
}

// What one run saves, the next run reads (FR-807): the forecast, its headers, its age and the spell.
func TestTheCacheSurvivesARestart(t *testing.T) {
	t.Parallel()
	settingsDir := t.TempDir()
	if err := New(settingsDir, &bytes.Buffer{}).Save(london, sample()); err != nil {
		t.Fatal(err)
	}
	loaded, err := New(settingsDir, &bytes.Buffer{}).Load()
	if err != nil || len(loaded) != 1 {
		t.Fatalf("loaded %d entries, %v; want one", len(loaded), err)
	}
	assertSame(t, loaded[london], sample())
}

// An entry that cannot be read is removed and left out, so its city is fetched again; the others
// load (FR-807).
func TestAnUnreadableCacheEntryIsFetchedAgain(t *testing.T) {
	t.Parallel()
	c, log := newCache(t)
	if err := c.Save(london, sample()); err != nil {
		t.Fatal(err)
	}
	put(t, c, oslo, `{"version":`)
	loaded, err := c.Load()
	if err != nil || len(loaded) != 1 || !slices.Equal(listing(t, c), []string{nameOf(london)}) {
		t.Errorf("loaded %v, %v; the folder holds %v", loaded, err, listing(t, c))
	}
	if !strings.Contains(log.String(), "cache "+oslo+": discarded: ") {
		t.Errorf("logged %q", log)
	}
}

// An entry larger than any forecast makes is discarded without being read whole.
func TestAnEntryTooLargeIsDiscarded(t *testing.T) {
	t.Parallel()
	c, log := newCache(t)
	put(t, c, oslo, strings.Repeat(" ", maxEntry+1))
	if loaded, err := c.Load(); err != nil || len(loaded) != 0 {
		t.Errorf("loaded %v, %v", loaded, err)
	}
	if !strings.Contains(log.String(), errTooLarge.Error()) {
		t.Errorf("logged %q", log)
	}
}

// No folder yet is no entries, not a fault.
func TestNoFolderIsNoEntries(t *testing.T) {
	t.Parallel()
	c, _ := newCache(t)
	if loaded, err := c.Load(); err != nil || len(loaded) != 0 {
		t.Errorf("loaded %v, %v", loaded, err)
	}
}

// A folder that cannot be listed for any reason but absence is a fault. A NUL byte in its path is
// refused on every platform. A file standing where the folder should be is not this case: Windows
// answers that it is not there.
func TestAFolderThatCannotBeListedIsAFault(t *testing.T) {
	t.Parallel()
	c := New(t.TempDir()+"\x00", &bytes.Buffer{})
	if _, err := c.Load(); err == nil || !strings.Contains(err.Error(), "listing") {
		t.Errorf("answered %v; want the fault", err)
	}
}

// A file whose name no id gives is not the cache's: it is neither read nor removed.
func TestFilesNotTheCachesAreLeftAlone(t *testing.T) {
	t.Parallel()
	c, log := newCache(t)
	put(t, c, london, "")
	if err := os.Remove(filepath.Join(c.dir, nameOf(london))); err != nil {
		t.Fatal(err)
	}
	foreign := []string{"notes.txt", "zz.json", "4A.json", ".json"}
	for _, name := range foreign {
		if err := os.WriteFile(filepath.Join(c.dir, name), []byte("x"), fileMode); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(c.dir, nameOf(oslo)), folderMode); err != nil {
		t.Fatal(err)
	}
	if loaded, err := c.Load(); err != nil || len(loaded) != 0 || log.Len() != 0 {
		t.Errorf("loaded %v, %v; logged %q", loaded, err, log)
	}
	if got := listing(t, c); len(got) != len(foreign)+1 {
		t.Errorf("the folder holds %v", got)
	}
}

// Ids a hand-edited settings file may hold, case and all, each keep a file of their own.
func TestEveryIDMakesASafeNameOfItsOwn(t *testing.T) {
	t.Parallel()
	c, _ := newCache(t)
	ids := []string{"ab", "AB", `..\..\x`, "a/b:c*?", strings.Repeat("z", MaxIDLength())}
	for _, id := range ids {
		if err := c.Save(id, sample()); err != nil {
			t.Fatalf("%q: %v", id, err)
		}
	}
	loaded, err := c.Load()
	if err != nil || len(loaded) != len(ids) {
		t.Errorf("loaded %d of %d, %v", len(loaded), len(ids), err)
	}
}

// An id that names no file is refused on every write.
func TestAnIDThatNamesNoFileIsRefused(t *testing.T) {
	t.Parallel()
	c, _ := newCache(t)
	for id, want := range map[string]error{"": ErrNoID, strings.Repeat("z", MaxIDLength()+1): ErrIDTooLong} {
		if err := c.Save(id, sample()); !errors.Is(err, want) {
			t.Errorf("save answered %v; want %v", err, want)
		}
		if err := c.Forget(id); !errors.Is(err, want) {
			t.Errorf("forget answered %v; want %v", err, want)
		}
	}
}

// A save that cannot make the folder fails, as does one with a figure JSON cannot hold.
func TestASaveThatCannotBeWrittenFails(t *testing.T) {
	t.Parallel()
	settingsFile := filepath.Join(t.TempDir(), "settings")
	if err := os.WriteFile(settingsFile, nil, fileMode); err != nil {
		t.Fatal(err)
	}
	if err := New(settingsFile, &bytes.Buffer{}).Save(london, sample()); err == nil || !strings.Contains(err.Error(), "making") {
		t.Errorf("answered %v; want the folder refused", err)
	}
	c, _ := newCache(t)
	bad := sample()
	bad.Forecast = unwritable()
	if err := c.Save(london, bad); err == nil {
		t.Error("a forecast JSON cannot hold was saved")
	}
}

// Forgetting removes the entry; forgetting one already gone is no fault, removing what cannot be
// removed is.
func TestForgettingRemovesTheEntry(t *testing.T) {
	t.Parallel()
	c, _ := newCache(t)
	if err := c.Save(london, sample()); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := c.Forget(london); err != nil {
			t.Fatal(err)
		}
	}
	if got := listing(t, c); len(got) != 0 {
		t.Errorf("the folder holds %v", got)
	}
	blocking := filepath.Join(c.dir, nameOf(oslo))
	if err := os.MkdirAll(filepath.Join(blocking, "inside"), folderMode); err != nil {
		t.Fatal(err)
	}
	if err := c.Forget(oslo); err == nil || !strings.Contains(err.Error(), "forgetting") {
		t.Errorf("answered %v; want the fault", err)
	}
}
