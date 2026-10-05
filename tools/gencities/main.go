// Command gencities writes the city list the place search reads (DATA-1, FR-202) from three GeoNames
// downloads: cities15000.zip, admin1CodesASCII.txt and countryInfo.txt, from
// https://download.geonames.org/export/dump/.
//
//	go run ./tools/gencities -dir C:\path\to\downloads -downloaded 2026-10-05
//
// The output is committed and embedded, so the application never reaches GeoNames. Each row keeps
// GeoNames' order, which settles ties in the search. The date the files were downloaded is written at
// its head.
package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// output is where the city list is written, relative to the repository root.
var output = filepath.Join("internal", "infrastructure", "places", "cities.tsv")

// The downloads read.
const (
	citiesArchive = "cities15000.zip"
	citiesFile    = "cities15000.txt"
	regionsFile   = "admin1CodesASCII.txt"
	countriesFile = "countryInfo.txt"
)

// dateLayout is how the download date is given and written.
const dateLayout = "2006-01-02"

// commentPrefix begins a line of countryInfo.txt that is not a country; it also begins the list's own
// head line.
const commentPrefix = "#"

// Columns of cities15000.txt, as GeoNames' readme numbers them.
const (
	cityID         = 0
	cityName       = 1
	cityASCIIName  = 2
	cityLatitude   = 4
	cityLongitude  = 5
	cityCountry    = 8
	cityRegionCode = 10
	cityPopulation = 14
	cityZone       = 17
	cityColumns    = 19
)

// Columns of admin1CodesASCII.txt ("GB.ENG", "England", ...) and of countryInfo.txt ("GB", ...,
// "United Kingdom" fifth).
const (
	regionKey      = 0
	regionName     = 1
	regionColumns  = 4
	countryKey     = 0
	countryName    = 4
	countryColumns = 19
)

// regionKeySeparator joins a country code and a region code into admin1CodesASCII.txt's key.
const regionKeySeparator = "."

// errMalformed is answered for a row with the wrong number of columns.
var errMalformed = errors.New("the row does not have the columns GeoNames documents")

func main() {
	dir := flag.String("dir", "", "folder holding cities15000.zip, admin1CodesASCII.txt and countryInfo.txt")
	downloaded := flag.String("downloaded", "", "the date the files were downloaded, as 2006-01-02")
	flag.Parse()
	if err := run(*dir, *downloaded, output); err != nil {
		fmt.Fprintln(os.Stderr, "gencities:", err)
		os.Exit(1)
	}
}

// run reads the three downloads in dir and writes the city list to path, stating downloaded.
func run(dir, downloaded, path string) error {
	if _, err := time.Parse(dateLayout, downloaded); err != nil {
		return fmt.Errorf("the download date %q is not a date like %s", downloaded, dateLayout)
	}
	regions, err := readTable(filepath.Join(dir, regionsFile), regionColumns, regionKey, regionName)
	if err != nil {
		return err
	}
	countries, err := readTable(filepath.Join(dir, countriesFile), countryColumns, countryKey, countryName)
	if err != nil {
		return err
	}
	archive, err := zip.OpenReader(filepath.Join(dir, citiesArchive))
	if err != nil {
		return err
	}
	defer archive.Close()
	cities, err := archive.Open(citiesFile)
	if err != nil {
		return fmt.Errorf("%s: %w", citiesArchive, err)
	}
	defer cities.Close()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	written, err := writeList(cities, regions, countries, downloaded, file)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	fmt.Printf("gencities: %d places written to %s\n", written, path)
	return nil
}

// readTable answers the file at path as key to name, passing over comment lines.
func readTable(path string, columns, key, name int) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return parseTable(file, filepath.Base(path), columns, key, name)
}

// parseTable reads a tab-separated table as key to name, passing over comment lines.
func parseTable(r io.Reader, source string, columns, key, name int) (map[string]string, error) {
	table := map[string]string{}
	scanner := bufio.NewScanner(r)
	for line := 1; scanner.Scan(); line++ {
		text := scanner.Text()
		if text == "" || strings.HasPrefix(text, commentPrefix) {
			continue
		}
		fields := strings.Split(text, "\t")
		if len(fields) != columns {
			return nil, fmt.Errorf("%s line %d: %w", source, line, errMalformed)
		}
		table[fields[key]] = fields[name]
	}
	return table, scanner.Err()
}

// writeList writes the head line stating downloaded, then one row per city: GeoNames id, name, ASCII
// name, region, country, latitude, longitude, population and zone, as GeoNames gives them. A region
// GeoNames names none for is empty; a city with no country name or no zone is refused, so a change in
// the downloads is seen rather than shipped. It answers how many places were written.
func writeList(cities io.Reader, regions, countries map[string]string, downloaded string, w io.Writer) (int, error) {
	out := bufio.NewWriter(w)
	fmt.Fprintf(out, "%s GeoNames %s, %s and %s, downloaded %s; licensed CC BY 4.0 by GeoNames\n",
		commentPrefix, citiesFile, regionsFile, countriesFile, downloaded)
	scanner := bufio.NewScanner(cities)
	written := 0
	for line := 1; scanner.Scan(); line++ {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) != cityColumns {
			return written, fmt.Errorf("%s line %d: %w", citiesFile, line, errMalformed)
		}
		country, ok := countries[fields[cityCountry]]
		if !ok || fields[cityZone] == "" {
			return written, fmt.Errorf("%s line %d: no country name for %q or no zone", citiesFile, line, fields[cityCountry])
		}
		region := regions[fields[cityCountry]+regionKeySeparator+fields[cityRegionCode]]
		row := []string{
			fields[cityID], fields[cityName], fields[cityASCIIName], region, country,
			fields[cityLatitude], fields[cityLongitude], fields[cityPopulation], fields[cityZone],
		}
		fmt.Fprintln(out, strings.Join(row, "\t"))
		written++
	}
	if err := scanner.Err(); err != nil {
		return written, err
	}
	return written, out.Flush()
}
