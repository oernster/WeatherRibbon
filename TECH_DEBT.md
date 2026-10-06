# Technical debt

What is still open, what is deliberately left and what only looks like debt.

Every item is a behaviour-preserving internal concern; nothing here reverts a feature or changes what
the user sees. `ARCHITECTURE.md` and the structural tests are the authority on the invariants.

Open items are numbered sections, so a scan for `## <number>.` tells whether the file is clear. The
two standing sections at the end are unnumbered. A resolved item is deleted outright; history belongs
in the release notes. Debt in the ribbon's own code is recorded where that code lives, in
ribbonkit's TECH_DEBT.md.

There is no open technical debt.

## Looks like debt, not worth touching

Nothing of WeatherRibbon's own; the ribbon's are in ribbonkit's TECH_DEBT.md.

## Not debt (do not "fix" these)

**The setup program holds no install logic.** WeatherRibbon's `installer` is only its composition
root: it carries the payload and the pictures, then hands them to the kit's setup window, which hands
every act to the kit's install policy.

**The window's port answers that there is no pull out.** WeatherRibbon has no sun map; the kit's
pull out is TimeRibbon's, so the adapter refusing one is the design, not a missing feature.

**The macOS and Linux build scripts are TimeRibbon's with the names changed.** They are shared in
shape on purpose, so a fix found in one is carried to the other; that they have not yet been run for
WeatherRibbon is a release check, not debt.
