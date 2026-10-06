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
