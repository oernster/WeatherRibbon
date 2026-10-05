package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixtures: two regions, two countries and three cities in GeoNames' shapes, one city with a
// region GeoNames does not name.
const (
	regionsText   = "GB.ENG\tEngland\tEngland\t6269131\nCA.08\tOntario\tOntario\t6093943\n"
	countriesText = "#ISO\tISO3\tISO-Numeric\tfips\tCountry\t5\t6\t7\t8\t9\t10\t11\t12\t13\t14\t15\t16\t17\t18\n" +
		"GB\tGBR\t826\tUK\tUnited Kingdom\t5\t6\t7\t8\t9\t10\t11\t12\t13\t14\t15\t16\t17\t18\n" +
		"CA\tCAN\t124\tCA\tCanada\t5\t6\t7\t8\t9\t10\t11\t12\t13\t14\t15\t16\t17\t18\n"
	downloaded = "2026-10-05"
)

// city answers a cities15000.txt row with the fields given and the rest filled.
func city(id, name, ascii, lat, lon, country, region, population, zone string) string {
	fields := make([]string, cityColumns)
	fields[cityID], fields[cityName], fields[cityASCIIName] = id, name, ascii
	fields[cityLatitude], fields[cityLongitude], fields[cityCountry] = lat, lon, country
	fields[cityRegionCode], fields[cityPopulation], fields[cityZone] = region, population, zone
	return strings.Join(fields, "\t")
}

// threeCities is London (England), London (Ontario) and a city whose region GeoNames does not name.
var threeCities = strings.Join([]string{
	city("2643743", "London", "London", "51.50853", "-0.12574", "GB", "ENG", "8961989", "Europe/London"),
	city("6058560", "London", "London", "42.98339", "-81.23304", "CA", "08", "346765", "America/Toronto"),
	city("1", "Nowhere", "Nowhere", "1", "2", "GB", "ZZZ", "15000", "Europe/London"),
}, "\n") + "\n"

// downloads writes the three downloads to a fresh folder, the cities file zipped under name.
func downloads(t *testing.T, cities, name string) string {
	t.Helper()
	dir := t.TempDir()
	for file, text := range map[string]string{regionsFile: regionsText, countriesFile: countriesText} {
		if err := os.WriteFile(filepath.Join(dir, file), []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	archive, err := os.Create(filepath.Join(dir, citiesArchive))
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(archive)
	entry, err := writer.Create(name)
	if err == nil {
		_, err = entry.Write([]byte(cities))
	}
	if err := errors.Join(err, writer.Close(), archive.Close()); err != nil {
		t.Fatal(err)
	}
	return dir
}

// The list holds the head line with the date, then each city's nine fields in GeoNames' order, the
// region named where GeoNames names it.
func TestTheListHoldsEachCityInOrder(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "cities.tsv")
	if err := run(downloads(t, threeCities, citiesFile), downloaded, path); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# GeoNames cities15000.txt, admin1CodesASCII.txt and countryInfo.txt, downloaded 2026-10-05; licensed CC BY 4.0 by GeoNames\n" +
		"2643743\tLondon\tLondon\tEngland\tUnited Kingdom\t51.50853\t-0.12574\t8961989\tEurope/London\n" +
		"6058560\tLondon\tLondon\tOntario\tCanada\t42.98339\t-81.23304\t346765\tAmerica/Toronto\n" +
		"1\tNowhere\tNowhere\t\tUnited Kingdom\t1\t2\t15000\tEurope/London\n"
	if string(written) != want {
		t.Errorf("wrote\n%s\nwant\n%s", written, want)
	}
}

// Every way the downloads or the date can be wrong is refused with what was wrong.
func TestWhatCannotBeReadIsRefused(t *testing.T) {
	t.Parallel()
	good := downloads(t, threeCities, citiesFile)
	out := filepath.Join(t.TempDir(), "cities.tsv")
	for name, tc := range map[string]struct {
		dir, date, path, want string
	}{
		"no date":            {good, "5 October", out, "is not a date"},
		"no regions":         {t.TempDir(), downloaded, out, regionsFile},
		"no cities archive":  {withoutArchive(t, good), downloaded, out, citiesArchive},
		"no cities file":     {downloads(t, threeCities, "other.txt"), downloaded, out, citiesFile},
		"a short city row":   {downloads(t, "1\tLondon\n", citiesFile), downloaded, out, "columns GeoNames documents"},
		"an unknown country": {downloads(t, city("1", "x", "x", "1", "2", "XX", "", "1", "Etc/UTC")+"\n", citiesFile), downloaded, out, "no country name"},
		"no zone":            {downloads(t, city("1", "x", "x", "1", "2", "GB", "", "1", "")+"\n", citiesFile), downloaded, out, "no zone"},
		"a line past a scan": {downloads(t, strings.Repeat("x", bufio.MaxScanTokenSize+1)+"\n", citiesFile), downloaded, out, "too long"},
		"nowhere to write":   {good, downloaded, filepath.Join(t.TempDir(), "missing", "cities.tsv"), "cities.tsv"},
	} {
		if err := run(tc.dir, tc.date, tc.path); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: answered %v; want %q", name, err, tc.want)
		}
	}
}

// A table with the wrong number of columns is refused, naming the file and line; comments and blank
// lines are passed over.
func TestAMalformedTableIsRefused(t *testing.T) {
	t.Parallel()
	table, err := parseTable(strings.NewReader("# note\n\nGB.ENG\tEngland\tEngland\t1\n"), regionsFile, regionColumns, regionKey, regionName)
	if err != nil || table["GB.ENG"] != "England" {
		t.Errorf("read %v (%v)", table, err)
	}
	if _, err := parseTable(strings.NewReader("GB.ENG\tEngland\n"), regionsFile, regionColumns, regionKey, regionName); !errors.Is(err, errMalformed) ||
		!strings.Contains(err.Error(), regionsFile+" line 1") {
		t.Errorf("answered %v", err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, regionsFile), []byte(regionsText), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, countriesFile), []byte("GB\tshort\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(dir, downloaded, filepath.Join(dir, "out.tsv")); !errors.Is(err, errMalformed) {
		t.Errorf("a short country row answered %v", err)
	}
}

// withoutArchive answers a copy of dir's two tables with no cities archive beside them.
func withoutArchive(t *testing.T, dir string) string {
	t.Helper()
	copyDir := t.TempDir()
	for _, file := range []string{regionsFile, countriesFile} {
		raw, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(copyDir, file), raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return copyDir
}
