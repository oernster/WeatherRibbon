package store

// NFR-C-1, the settings file contract of 1.0.0: every later release reads every file 1.0.0 writes to
// the same settings. testdata/settings-1.0.0.json is a file in 1.0.0's shape with every key set away
// from its default, so a version that stops reading any one of them fails here. The fixture is
// frozen: it is never regenerated from a later writer, since what it proves is that the old shape
// still reads. A later version may add keys; it may not stop reading or change these.

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// contractFixture is a settings file as 1.0.0 writes it.
var contractFixture = filepath.Join("testdata", "settings-1.0.0.json")

func TestA1Point0SettingsFileIsReadWhole(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(contractFixture)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	write(t, dir, string(raw))
	store := New(dir, productName)
	loaded, err := store.Load()
	if err != nil || loaded.Notice != "" {
		t.Fatalf("a 1.0.0 file did not load cleanly: %v %q", err, loaded.Notice)
	}
	if extras := store.file.Extras(); len(extras) != 0 {
		t.Errorf("a 1.0.0 key is no longer read, only carried: %v", extras)
	}
	if got, want := loaded.Settings, full(); !reflect.DeepEqual(got, want) {
		t.Errorf("a 1.0.0 file read as\n%+v\nwant\n%+v", got, want)
	}
}

// The fixture is what this version writes for the same settings, so it is 1.0.0's shape exactly. Once
// 1.0.0 ships this test is deleted and the fixture kept as it is.
func TestTheFixtureIsWhatThisVersionWrites(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(contractFixture)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := New(dir, productName).Save(full()); err != nil {
		t.Fatal(err)
	}
	// A checkout may turn the fixture's line ends into CRLF; the shape is what is compared.
	if written := read(t, dir); written != strings.ReplaceAll(string(raw), "\r\n", "\n") {
		t.Errorf("this version writes\n%s\nthe fixture holds\n%s", written, raw)
	}
}
