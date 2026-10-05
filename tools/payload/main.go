//go:build windows

// Command payload packs the setup program's payload: every file of the built application, then the
// licence beside it. The work is the kit's delivery.Payload; this names WeatherRibbon.
//
//	go run ./tools/payload -app build/bin -licence LICENSE -out installer/payload.zip
package main

import (
	"fmt"
	"os"

	"github.com/oernster/ribbonkit/infrastructure/delivery"
	"github.com/oernster/ribbonkit/infrastructure/setup"
	"github.com/oernster/weatherribbon/internal/product"
)

func main() {
	if err := delivery.Payload(os.Args[1:], os.Stdout, setup.Product{App: product.App()}); err != nil {
		fmt.Fprintln(os.Stderr, "payload:", err)
		os.Exit(1)
	}
}
