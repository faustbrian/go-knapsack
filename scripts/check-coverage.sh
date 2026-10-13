#!/bin/sh
set -eu

# The same strict owner acceptance runs locally and from package-test.
profile="$(mktemp "${TMPDIR:-/tmp}/knapsack-coverage.XXXXXX")"
trap 'rm -f "$profile"' EXIT HUP INT TERM
packages="$(go run scripts/coverage-exception/check.go --targets)"
go test ./... -covermode=atomic -coverpkg="$packages" -coverprofile="$profile" -count=1
go run scripts/coverage-exception/check.go --profile "$profile"
