#!/usr/bin/env bash

set -euo pipefail

# Write bin/SHA256SUMS and upload every file in bin/ to the GitHub release for the tag.
# If the tag has no release yet, create a draft release.
#
# Required environment variables:
#   GH_TOKEN    Token with contents: write on the repository
#   GH_REPO     The repository, as owner/name
#
# Usage: publish-release.sh <tag>

# shellcheck source=sha256.sh
source "$(dirname "${BASH_SOURCE[0]}")/sha256.sh"

function main {
	local -r tag="$1"

	(cd bin && sha256 -- * >SHA256SUMS)

	if ! gh release view "$tag" >/dev/null 2>&1; then
		gh release create "$tag" --verify-tag --draft --title "$tag" --notes ""
	fi
	gh release upload "$tag" bin/* --clobber
}

main "$@"
