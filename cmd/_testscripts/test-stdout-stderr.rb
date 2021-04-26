#!/usr/bin/env ruby
# This scripts writes some text to stdout and stderr and then exits with an error. This is used to test that git-xargs
# always logs the stdout and stderr from scripts, even if those scripts exit with an error.

STDOUT.puts('Hello, from STDOUT')
STDERR.puts('Hello, from STDERR')
exit 1