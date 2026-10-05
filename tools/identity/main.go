// Command identity prints the names the Linux and macOS builds need, as shell assignments read from
// the one home each has in internal/product. The work is the kit's delivery.ShellNames.
//
//	eval "$(go run ./tools/identity)"
package main

import (
	"os"

	"github.com/oernster/ribbonkit/infrastructure/delivery"
	"github.com/oernster/weatherribbon/internal/product"
)

func main() { delivery.ShellNames(os.Stdout, product.App(), product.Copyright) }
