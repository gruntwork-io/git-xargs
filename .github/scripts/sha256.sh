# shellcheck shell=bash

# SHA-256 helpers for the other scripts in this directory. Source this file; don't run it.

# Print "<hash>  <file>" for each file, with sha256sum or shasum, whichever the host has.
function sha256 {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$@"
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$@"
	else
		echo "ERROR: Neither sha256sum nor shasum is installed." >&2
		return 1
	fi
}

# Fail unless the file's SHA-256 is the expected hash.
#
# Usage: verify_sha256 <expected-hash> <file>
function verify_sha256 {
	local -r expected="$1"
	local -r file="$2"
	local actual
	actual="$(sha256 -- "$file")"
	actual="${actual%% *}"

	if [[ $actual != "$expected" ]]; then
		echo "ERROR: $file has SHA-256 $actual, expected $expected." >&2
		return 1
	fi
	echo "$file: SHA-256 OK"
}
