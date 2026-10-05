package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
)

// weatherIcons is the Yr icon set the page draws, embedded here only for its file names: the files
// are the one home of which symbols have an icon (FR-412).
//
//go:embed frontend/src/assets/weather/*.svg
var weatherIcons embed.FS

// iconPattern finds every icon in weatherIcons; iconExtension is each one's extension.
const (
	iconPattern   = "frontend/src/assets/weather/*.svg"
	iconExtension = ".svg"
)

// iconSet is the application's Icons port over the embedded set, telling the log of each symbol
// that has no icon.
type iconSet struct {
	names map[string]bool
	log   io.Writer
}

// newIconSet answers the icon set held in files under pattern, logging to log.
func newIconSet(files fs.FS, pattern string, log io.Writer) (iconSet, error) {
	found, err := fs.Glob(files, pattern)
	if err != nil {
		return iconSet{}, fmt.Errorf("listing the weather icons: %w", err)
	}
	names := make(map[string]bool, len(found))
	for _, file := range found {
		names[strings.TrimSuffix(path.Base(file), iconExtension)] = true
	}
	return iconSet{names: names, log: log}, nil
}

// Has answers whether the set holds an icon named name.
func (s iconSet) Has(name string) bool { return s.names[name] }

// Missing logs a symbol code the set holds no icon for; the application tells each code once.
func (s iconSet) Missing(code string) {
	fmt.Fprintf(s.log, "no weather icon for the symbol %q; its words are shown instead\n", code)
}
