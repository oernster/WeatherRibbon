"""Generate WeatherRibbon's icons from the master artwork in assets/.

The work is ribbonkit's tools/genicons.py, read from the kit Go builds against; this names
WeatherRibbon's button artwork with the height the page draws each at. Its docstring lists every
file written. Run it when a master changes:

    python tools/genicons.py

It is not part of the build. The output is committed, so a clone needs neither Python nor Pillow
to build.
"""

from __future__ import annotations

import importlib.util
import pathlib
import subprocess
import sys

REPO = pathlib.Path(__file__).resolve().parent.parent
KIT_MODULE = "github.com/oernster/ribbonkit"

# BUTTONS names each button master with the height the page draws it at: the donate mark at the foot
# of Settings, the Add city button's artwork drawn large.
BUTTONS = {"donate.png": 32, "add-city.png": 128}


def kit_genicons():
    """Load the kit's genicons module from where Go builds the kit."""
    found = subprocess.run(
        ["go", "list", "-m", "-f", "{{.Dir}}", KIT_MODULE], cwd=REPO, capture_output=True, text=True
    )
    if found.returncode != 0 or not found.stdout.strip():
        sys.exit(f"go list could not find {KIT_MODULE}: {found.stderr.strip()}; run go mod download")
    path = pathlib.Path(found.stdout.strip()) / "tools" / "genicons.py"
    spec = importlib.util.spec_from_file_location("ribbonkit_genicons", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


if __name__ == "__main__":
    raise SystemExit(kit_genicons().generate(REPO, BUTTONS))
