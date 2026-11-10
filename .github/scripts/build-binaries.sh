#!/usr/bin/env bash

set -euo pipefail

# Cross-compile git-xargs into bin/ for every platform that a release ships.
#
# Usage: build-binaries.sh [version]

readonly TARGETS=(darwin/amd64 darwin/arm64 linux/386 linux/amd64 linux/arm64 windows/386 windows/amd64)

function main {
	local -r version="${1:-}"
	local target os arch ext

	for target in "${TARGETS[@]}"; do
		os="${target%/*}"
		arch="${target#*/}"
		ext=""
		if [[ $os == "windows" ]]; then
			ext=".exe"
		fi

		echo "Building git-xargs_${os}_${arch}${ext}"
		CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
			go build -ldflags "-X main.VERSION=${version}" -o "bin/git-xargs_${os}_${arch}${ext}" .
	done
}

main "$@"
