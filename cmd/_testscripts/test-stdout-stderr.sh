#!/usr/bin/env bash
# This script writes some text to stdout and stderr and then exits with an error. This is used to test that git-xargs
# always logs the stdout and stderr from scripts, even if those scripts exit with an error.

echo 'Hello, from STDOUT'
>&2 echo 'Hello, from STDERR'
exit 1