#!/usr/bin/env bash
# This script writes some text to stdout and stderr and then exits.
# This is used to test that git-xargs registers environment variables based on flags and arguments.

echo "XARGS_REPO_NAME=$XARGS_REPO_NAME"
echo "XARGS_DRY_RUN=$XARGS_DRY_RUN"