#!/usr/bin/env bash

set -euo pipefail

# Install gon, which signs and notarizes the macOS binaries.
#
# Usage: install-gon.sh <gon-version> <gon_macos.zip-sha256>

# shellcheck source=sha256.sh
source "$(dirname "${BASH_SOURCE[0]}")/sha256.sh"

function main {
	local -r version="$1"
	local -r sha256="$2"
	local tmp_dir
	tmp_dir="$(mktemp -d)"

	curl -fsSL -o "$tmp_dir/gon.zip" "https://github.com/Bearer/gon/releases/download/${version}/gon_macos.zip"
	verify_sha256 "$sha256" "$tmp_dir/gon.zip"
	unzip -o "$tmp_dir/gon.zip" gon -d "$tmp_dir"
	sudo mv "$tmp_dir/gon" /usr/local/bin/gon
	rm -rf "$tmp_dir"

	gon --version
}

main "$@"
