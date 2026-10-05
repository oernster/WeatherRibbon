# Builds WeatherRibbon: the application and its setup program.
#
#   ./build.ps1                 build the application and the setup program
#   ./build.ps1 -SkipInstaller  build only the application
#
# Outputs:
#   build/bin/WeatherRibbon.exe            the application
#   dist-installer/WeatherRibbonSetup.exe  the setup program, carrying the application; named by
#                                      outputfilename in installer/wails.json, with no version
#
# The version is read from VERSION and stamped into both executables with -ldflags -X, so no
# version literal lives anywhere in the source (CON-4). -X reaches only a var; against a const it
# silently does nothing, which is why product.Version is a var.
param(
    [switch]$SkipInstaller
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

$version = (Get-Content (Join-Path $root 'VERSION') -Raw).Trim()
if ($version -notmatch '^[0-9]+\.[0-9]+\.[0-9]+$') { throw "VERSION holds '$version', not major.minor.patch" }

# The names come from where they are already stated: the module from go.mod, each executable's name
# from its wails.json.
$module = go list -m
if ($LASTEXITCODE -ne 0) { throw "go list -m failed with exit code $LASTEXITCODE" }
$appName = (Get-Content (Join-Path $root 'wails.json') -Raw | ConvertFrom-Json).outputfilename
$setupName = (Get-Content (Join-Path $root 'installer/wails.json') -Raw | ConvertFrom-Json).outputfilename
$ldflags = "-X $module/internal/product.Version=$version"
Write-Host "Building $appName $version"

# Cgo is pinned off rather than left to whatever the machine defaults to: nothing here wants a C
# toolchain; the first machine with one would otherwise quietly build a different binary. It is
# set before the gate as well, so the tests exercise the configuration that ships.
$env:CGO_ENABLED = '0'

# What ships is built from the ribbonkit tag go.mod requires, never from a working copy a local
# go.work points at while the kit is being changed beside WeatherRibbon.
$env:GOWORK = 'off'

# Verify before building, with no way past it (NFR-M-1). A gate that can be skipped is skipped on
# the day it would have caught something: run test.ps1 directly while working; let the build insist.
& (Join-Path $root 'test.ps1')
if ($LASTEXITCODE -ne 0) { throw "test.ps1 failed with exit code $LASTEXITCODE" }

# One icon is the whole identity: both executables wear build/windows/icon.ico, which
# tools/genicons.py makes from assets/application-icon.png and which is committed. The setup
# program's build folder is output, so the icon is copied in before each build rather than kept
# there as a second copy.
$icon = Join-Path $root 'build/windows/icon.ico'
$appIcon = Join-Path $root 'build/appicon.png'
if (-not ((Test-Path $icon) -and (Test-Path $appIcon))) {
    throw 'Missing build/windows/icon.ico or build/appicon.png: run python tools/genicons.py first.'
}

# Each executable's Windows version resource is written from VERSION and internal/product before
# its build. Left to Wails' template it carried Wails' fallback version and a placeholder copyright.
Write-Host 'Writing the version resources...'
go run ./tools/versioninfo -version $version -out (Join-Path $root 'build/windows/info.json')
if ($LASTEXITCODE -ne 0) { throw "tools/versioninfo failed with exit code $LASTEXITCODE" }

Write-Host 'Building the application...'
wails build -ldflags $ldflags
if ($LASTEXITCODE -ne 0) { throw "wails build failed with exit code $LASTEXITCODE" }

if ($SkipInstaller) {
    Write-Host "Done: build/bin/$appName.exe ($version)"
    exit 0
}

Write-Host 'Packing the application as the setup payload...'
$payload = Join-Path $root 'installer/payload.zip'
go run ./tools/payload -app (Join-Path $root 'build/bin') -licence (Join-Path $root 'LICENSE') -out $payload
if ($LASTEXITCODE -ne 0) { throw "tools/payload failed with exit code $LASTEXITCODE" }

try {
    Write-Host 'Building the setup program...'
    $setupBuild = Join-Path $root 'installer/build'
    New-Item -ItemType Directory -Force -Path (Join-Path $setupBuild 'windows') | Out-Null
    Copy-Item $icon (Join-Path $setupBuild 'windows/icon.ico') -Force
    Copy-Item $appIcon (Join-Path $setupBuild 'appicon.png') -Force
    go run ./tools/versioninfo -version $version -setup -out (Join-Path $setupBuild 'windows/info.json')
    if ($LASTEXITCODE -ne 0) { throw "tools/versioninfo failed with exit code $LASTEXITCODE" }
    Push-Location (Join-Path $root 'installer')
    try {
        wails build -ldflags $ldflags
        if ($LASTEXITCODE -ne 0) { throw "wails build of the setup program failed with exit code $LASTEXITCODE" }
    } finally {
        Pop-Location
    }

    $distDir = Join-Path $root 'dist-installer'
    New-Item -ItemType Directory -Force -Path $distDir | Out-Null
    $distributable = Join-Path $distDir "$setupName.exe"
    Copy-Item (Join-Path $setupBuild "bin/$setupName.exe") $distributable -Force
} finally {
    # The empty-zip placeholder goes back whether or not the setup program built, so go build and
    # the tests work without a full build and the real payload is never left where a commit could
    # pick it up. It is the 22-byte end-of-central-directory record of an archive with no entries.
    $empty = [byte[]](0x50, 0x4B, 0x05, 0x06) + (New-Object byte[] 18)
    [System.IO.File]::WriteAllBytes($payload, $empty)
}

$size = (Get-Item $distributable).Length
Write-Host ("Done: {0} ({1:N0} bytes)" -f $distributable, $size)
