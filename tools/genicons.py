"""Generate WeatherRibbon's icons from the master artwork in assets/, ported from TimeRibbon's.

What comes out, every file from one of the masters so the masters stay the one home:

  build/windows/icon.ico               the multi-size Windows icon the executable wears; the
                                       taskbar, the tray and the shortcuts read it out of the binary
  build/appicon.png                    the artwork Wails keeps beside it, which the tray on Linux
                                       and the menu bar on macOS are handed
  frontend/src/assets/donate.png       the donate mark at the foot of Settings
  frontend/src/assets/add-city.png     the Add city button's artwork
  frontend/src/assets/app-icon.png     the application icon at the head of About

The setup program's header and theme toggle are added here when the setup program is ported.

Run it when a master changes:

    python tools/genicons.py

It is not part of the build. The output is committed, so a clone needs neither Python nor Pillow
to build.
"""

from __future__ import annotations

import pathlib
import sys

try:
    from PIL import Image
except ImportError:  # pragma: no cover - a plain message beats a traceback
    sys.exit("Pillow is required: python -m pip install pillow")

REPO = pathlib.Path(__file__).resolve().parent.parent
MASTERS = REPO / "assets"
APP_MASTER = MASTERS / "application-icon.png"

# ICO_SIZES are the sizes Windows chooses between: the tray and menu sizes, the taskbar and
# shortcut sizes, then the large one Explorer uses in its biggest view. Leaving one out makes
# Windows scale a neighbour, which looks soft.
ICO_SIZES = [(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]

# APPICON_SIZE is the square Wails expects its build/appicon.png to be.
APPICON_SIZE = 1024

# Button artwork is drawn at a button's height, not as an icon: each is cropped to its artwork and
# scaled by height alone. BUTTON_MASTERS names each master with the height the page draws it at
# (.art img and .art.large img in the page's CSS); the render is RENDER_SCALE times that, so it
# stays crisp under display scaling.
RENDER_SCALE = 4
PAGE_ASSETS = REPO / "frontend" / "src" / "assets"
BUTTON_MASTERS = {"donate.png": 32, "add-city.png": 128}

# ABOUT_ICON_DRAWN is the application icon's size at the head of About; it is rendered
# RENDER_SCALE times that, square.
ABOUT_ICON_DRAWN = 128
ABOUT_ICON_TARGET = PAGE_ASSETS / "app-icon.png"

ICO_TARGET = REPO / "build" / "windows" / "icon.ico"
APPICON_TARGET = REPO / "build" / "appicon.png"


def squared(master: pathlib.Path) -> Image.Image:
    """Open a master, trim any transparent margin and centre it on a transparent square."""
    image = Image.open(master).convert("RGBA")
    box = image.getbbox()
    if box is not None:
        image = image.crop(box)
    side = max(image.width, image.height)
    square = Image.new("RGBA", (side, side), (0, 0, 0, 0))
    square.paste(image, ((side - image.width) // 2, (side - image.height) // 2), image)
    return square


def write_png(image: Image.Image, size: int, target: pathlib.Path) -> None:
    """Write image scaled to a size-pixel square and say what was written."""
    target.parent.mkdir(parents=True, exist_ok=True)
    image.resize((size, size), Image.LANCZOS).save(target, "PNG", optimize=True)
    print(f"{target.relative_to(REPO).as_posix():<42} {size:>5} px {target.stat().st_size:>9,} bytes")


def button_art(master: pathlib.Path, height: int) -> Image.Image:
    """Button artwork cropped to its pixels and scaled to height, keeping its shape."""
    image = Image.open(master).convert("RGBA")
    box = image.getbbox()
    if box is not None:
        image = image.crop(box)
    width = round(image.width * height / image.height)
    return image.resize((width, height), Image.LANCZOS)


def main() -> int:
    for master in (APP_MASTER, *(MASTERS / name for name in BUTTON_MASTERS)):
        if not master.exists():
            sys.exit(f"no master artwork at {master}")

    app = squared(APP_MASTER)
    ICO_TARGET.parent.mkdir(parents=True, exist_ok=True)
    app.save(ICO_TARGET, "ICO", sizes=ICO_SIZES)
    sizes = ", ".join(str(width) for width, _ in ICO_SIZES)
    print(f"{ICO_TARGET.relative_to(REPO).as_posix():<42} {sizes} {ICO_TARGET.stat().st_size:>9,} bytes")
    write_png(app, APPICON_SIZE, APPICON_TARGET)
    write_png(app, RENDER_SCALE * ABOUT_ICON_DRAWN, ABOUT_ICON_TARGET)
    for name, drawn in BUTTON_MASTERS.items():
        art = button_art(MASTERS / name, RENDER_SCALE * drawn)
        target = PAGE_ASSETS / name
        target.parent.mkdir(parents=True, exist_ok=True)
        art.save(target, "PNG", optimize=True)
        print(f"{target.relative_to(REPO).as_posix():<42} {art.width}x{art.height} {target.stat().st_size:>9,} bytes")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
