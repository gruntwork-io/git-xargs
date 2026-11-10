#!/usr/bin/env bash

set -euo pipefail

# Sign and notarize bin/git-xargs_darwin_{amd64,arm64} with gon, replacing the unsigned binaries.
#
# Required environment variables:
#   AC_PASSWORD                   Apple ID app-specific password, read by gon
#   AC_PROVIDER                   Apple team ID, read by gon
#   MACOS_CERTIFICATE             Base64-encoded .p12 Developer ID Application certificate
#   MACOS_CERTIFICATE_PASSWORD    Password for the .p12 file
#
# Usage: sign-macos-binaries.sh

# Apple's Developer ID G2 intermediate. It chains to the system-trusted Apple Root CA, so it needs no trust settings.
readonly DEVELOPER_ID_G2_URL="https://www.apple.com/certificateauthority/DeveloperIDG2CA.cer"
readonly DEVELOPER_ID_G2_SHA256="f16cd3c54c7f83cea4bf1a3e6a0819c8aaa8e4a1528fd144715f350643d2df3a"

function main {
	local var
	for var in AC_PASSWORD AC_PROVIDER MACOS_CERTIFICATE MACOS_CERTIFICATE_PASSWORD; do
		if [[ -z ${!var:-} ]]; then
			echo "::error::The secret for $var is not set."
			exit 1
		fi
	done

	local tmp_dir
	tmp_dir="$(mktemp -d)"

	import_certificates "$tmp_dir"

	local arch
	for arch in amd64 arm64; do
		gon -log-level=info ".gon_${arch}.hcl"
		unzip -o "git-xargs_darwin_${arch}.zip" -d bin
		codesign --verify --verbose "bin/git-xargs_darwin_${arch}"
	done

	rm -rf "$tmp_dir"
}

function import_certificates {
	local -r tmp_dir="$1"
	local -r keychain="$tmp_dir/signing.keychain-db"
	local keychain_password
	keychain_password="$(uuidgen)"

	security create-keychain -p "$keychain_password" "$keychain"
	security default-keychain -s "$keychain"
	security unlock-keychain -p "$keychain_password" "$keychain"

	echo "$MACOS_CERTIFICATE" | base64 -d >"$tmp_dir/certificate.p12"
	security import "$tmp_dir/certificate.p12" -f pkcs12 -k "$keychain" -P "$MACOS_CERTIFICATE_PASSWORD" -T /usr/bin/codesign
	rm "$tmp_dir/certificate.p12"

	curl -fsSL -o "$tmp_dir/DeveloperIDG2CA.cer" "$DEVELOPER_ID_G2_URL"
	echo "$DEVELOPER_ID_G2_SHA256  $tmp_dir/DeveloperIDG2CA.cer" | shasum -a 256 -c -
	security import "$tmp_dir/DeveloperIDG2CA.cer" -k "$keychain"

	security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k "$keychain_password" "$keychain"
}

main "$@"
