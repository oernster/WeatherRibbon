# Verifies WeatherRibbon: formatting, vet, staticcheck, the whole suite and the coverage floors.
#
#   ./test.ps1              run everything
#   ./test.ps1 -Floor 95    run with a different floor over domain and application
#
# build.ps1, once ported, runs this before it builds, so a release cannot be cut from a tree that fails it.
#
# Every floor is the measured number, not a target. A floor picked from an aspiration only teaches
# people to lower it; one at the measured number fails the moment cover is lost.
param(
    [double]$Floor = 100
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

# Domain and application are the layers a test reaches with no filesystem, clock or display, so
# anything short of 100 percent there is a decision nobody made (CON-3).
$gated = './internal/domain/...', './internal/application/...'

# The Go tools are pointed at this list rather than at ./..., which would reach into
# frontend/node_modules once the front end exists.
$packages = go list ./... | Where-Object { $_ -notmatch '/node_modules/' }
if ($LASTEXITCODE -ne 0) { throw "go list failed with exit code $LASTEXITCODE" }

Write-Host 'Checking formatting...'
$unformatted = gofmt -l . | Where-Object { $_ -notmatch '^frontend' }
if ($unformatted) { throw "gofmt reports unformatted files:`n$($unformatted -join "`n")" }

Write-Host 'Vetting...'
go vet $packages
if ($LASTEXITCODE -ne 0) { throw "go vet failed with exit code $LASTEXITCODE" }

# Pinned so a new release of the checker cannot fail a change that touched nothing it reads. Raise
# it on purpose, having read what the new version reports.
$staticcheckVersion = 'v0.8.1'
Write-Host "Running staticcheck $staticcheckVersion..."
go run "honnef.co/go/tools/cmd/staticcheck@$staticcheckVersion" $packages
if ($LASTEXITCODE -ne 0) { throw "staticcheck failed with exit code $LASTEXITCODE" }

Write-Host 'Running the whole suite...'
go test -count=1 $packages
if ($LASTEXITCODE -ne 0) { throw "go test failed with exit code $LASTEXITCODE" }

# The page is held to the same bar in its own runner. A missing node_modules stops the gate rather
# than skipping the page: a check that quietly does not run is the one that is not there on the day.
Write-Host 'Checking the front end...'
$frontend = Join-Path $root 'frontend'
if (-not (Test-Path (Join-Path $frontend 'node_modules'))) {
    throw "the front end's dependencies are not installed: run npm install in $frontend, then run this again"
}
Push-Location $frontend
try {
    foreach ($check in 'lint', 'typecheck', 'test') {
        npm run $check
        if ($LASTEXITCODE -ne 0) { throw "npm run $check failed with exit code $LASTEXITCODE" }
    }
} finally {
    Pop-Location
}

Write-Host "Measuring coverage of $($gated -join ', ')..."
$profilePath = Join-Path ([System.IO.Path]::GetTempPath()) 'weatherribbon-coverage.out'
try {
    go test -count=1 "-coverprofile=$profilePath" @gated
    if ($LASTEXITCODE -ne 0) { throw "the coverage run failed with exit code $LASTEXITCODE" }
    $summary = go tool cover "-func=$profilePath"
    if ($LASTEXITCODE -ne 0) { throw "go tool cover failed with exit code $LASTEXITCODE" }
    $total = ($summary | Select-Object -Last 1)
    if ($total -notmatch '([0-9]+(?:\.[0-9]+)?)%\s*$') { throw "could not read a total from: $total" }
    $percent = [double]$Matches[1]
    if ($percent -lt $Floor) {
        Write-Host 'Not covered:'
        $summary | Where-Object { $_ -notmatch '100\.0%\s*$' -and $_ -notmatch '^total:' } | ForEach-Object { Write-Host "  $_" }
        throw "coverage is $percent%, below the floor of $Floor%"
    }
    Write-Host "Coverage $percent%, floor $Floor%."
} finally {
    if (Test-Path $profilePath) { Remove-Item $profilePath -Force }
}

# The rest of the tree, each package held at the number measured when its floor was set. The root
# package's shortfall is the composition root (main.go). ribbonkit's packages are held by the kit's
# own gate.
$measured = [ordered]@{
    '.'                                  = 68
    './internal/infrastructure/cache'    = 100
    './internal/infrastructure/metno'    = 100
    './internal/infrastructure/pacing'   = 100
    './internal/infrastructure/places'   = 100
    './internal/infrastructure/store'    = 100
    './internal/product'                 = 100
    './tools/gencities'                  = 90
}

Write-Host 'Measuring the rest of the tree...'
foreach ($package in $measured.Keys) {
    $floorHere = $measured[$package]
    $reported = go test -count=1 -cover $package
    # The report is held to read the coverage from, so a failing test's own words would otherwise
    # never reach the log; the gate named the package once and nothing else (2026-10-05).
    if ($LASTEXITCODE -ne 0) {
        $reported | ForEach-Object { Write-Host $_ }
        throw "$package failed with exit code $LASTEXITCODE"
    }
    $line = $reported | Where-Object { $_ -match 'coverage: ' } | Select-Object -First 1
    if ($line -notmatch 'coverage: ([0-9]+(?:\.[0-9]+)?)%') { throw "could not read a coverage figure for ${package}: $line" }
    $reached = [double]$Matches[1]
    if ($reached -lt $floorHere) { throw "$package is at $reached%, below its floor of $floorHere%" }
    Write-Host ("  {0,-38} {1,5}%  floor {2}%" -f $package, $reached, $floorHere)
}

Write-Host 'All green.'
