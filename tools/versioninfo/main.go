// Command versioninfo writes the Windows version resource wails build puts into an executable, from
// the version build.ps1 read out of VERSION and the identity in internal/product. The work is the
// kit's delivery.VersionInfo; this names WeatherRibbon.
//
//	go run ./tools/versioninfo -version 1.0.0 -out build/windows/info.json
//	go run ./tools/versioninfo -version 1.0.0 -setup -out installer/build/windows/info.json
package main

import (
	"fmt"
	"os"

	"github.com/oernster/ribbonkit/infrastructure/delivery"
	"github.com/oernster/weatherribbon/internal/product"
)

func main() {
	about := delivery.About{Name: product.Name, Author: product.Author, Copyright: product.Copyright}
	if err := delivery.VersionInfo(os.Args[1:], os.Stderr, about); err != nil {
		fmt.Fprintln(os.Stderr, "versioninfo:", err)
		os.Exit(1)
	}
}
