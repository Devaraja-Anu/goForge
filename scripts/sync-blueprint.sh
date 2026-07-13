#!/usr/bin/env bash
set -euo pipefail

# Regenerates blueprintsrc/ from blueprint/ so it can be embedded via
# go:embed. blueprint/ cannot be embedded directly: it has its own go.mod,
# making it a separate Go module, and go:embed cannot see into another
# module's directory tree at all, regardless of embed pattern.
#
# blueprintsrc/ is a derived mirror, NOT a second source of truth.
# Never edit files inside it directly — edit blueprint/, then re-run this
# script. CI catches drift via `make check-blueprint-sync`.
#
# Every .go file (and go.mod/go.sum) is renamed with a .tmpl suffix.
# This is required, not cosmetic: without it, blueprintsrc/ has no go.mod
# of its own, which means `go build ./...` in GoForge's own module walks
# into blueprintsrc/ and tries to compile it as real GoForge source —
# failing to resolve blueprint's own dependency graph, since it was never
# meant to be part of GoForge's go.mod. Renaming to .tmpl removes it from
# the Go toolchain's view entirely while leaving it fully visible to
# go:embed, which is extension-agnostic. internal/generator reverses the
# rename when writing files into a newly generated project.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="$ROOT_DIR/blueprint"
DST="$ROOT_DIR/blueprintsrc"

rm -rf "$DST"
mkdir -p "$DST"

cp -R "$SRC"/. "$DST"/
rm -rf "$DST/.git"

mv "$DST/go.mod" "$DST/go.mod.tmpl"
mv "$DST/go.sum" "$DST/go.sum.tmpl"

# Rename every .go file to .go.tmpl, preserving directory structure.
find "$DST" -type f -name '*.go' | while IFS= read -r f; do
	mv "$f" "${f}.tmpl"
done

echo "blueprintsrc/ synced from blueprint/ ($(find "$DST" -name '*.go.tmpl' | wc -l) .go files renamed)"