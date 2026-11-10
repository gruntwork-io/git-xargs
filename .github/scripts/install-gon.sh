#!/usr/bin/env bash

set -euo pipefail

# Install gon, which signs and notarizes the macOS binaries.
#
# Usage: install-gon.sh <gon-version>

function main {
	local -r version="$1"
	local tmp_dir
	tmp_dir="$(mktemp -d)"

	curl -fsSL -o "$tmp_dir/gon.zip" "https://github.com/Bearer/gon/releases/download/${version}/gon_macos.zip"
	unzip -o "$tmp_dir/gon.zip" gon -d "$tmp_dir"
	sudo mv "$tmp_dir/gon" /usr/local/bin/gon
	rm -rf "$tmp_dir"

	gon --version
}

main "$@"
