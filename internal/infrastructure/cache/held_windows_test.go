package cache

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/infrastructure/heldfile"
)

// An entry another program holds open cannot be read, so it is left out and fetched again. It
// cannot be removed either; the log says so. The others load.
func TestAnEntryHeldOpenIsLeftOutAndSaysItWasNotRemoved(t *testing.T) {
	t.Parallel()
	c, log := newCache(t)
	for _, id := range []string{london, oslo} {
		if err := c.Save(id, sample()); err != nil {
			t.Fatal(err)
		}
	}
	release, err := heldfile.Hold(filepath.Join(c.dir, nameOf(oslo)))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := c.Load()
	release()
	if _, held := loaded[oslo]; err != nil || len(loaded) != 1 || held {
		t.Errorf("loaded %v, %v; want London alone", loaded, err)
	}
	if !strings.Contains(log.String(), "cache "+oslo+": discarded: ") || !strings.Contains(log.String(), "not removed") {
		t.Errorf("logged %q", log)
	}
}
