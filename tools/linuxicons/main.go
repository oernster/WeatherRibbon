// Command linuxicons writes the application's icon at each size the Linux icon theme asks for. The
// work is the kit's delivery.ThemeIcons.
//
//	go run ./tools/linuxicons -in build/appicon.png -out build/linux/icons -name weatherribbon
package main

import (
	"fmt"
	"os"

	"github.com/oernster/ribbonkit/infrastructure/delivery"
)

func main() {
	if err := delivery.ThemeIcons(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "linuxicons:", err)
		os.Exit(1)
	}
}
