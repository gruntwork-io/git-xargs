#!/usr/bin/env bash
# We renamed a ton of our repos to the Terraform Registry naming format. This script uses grep and sed to search for
# references to the old names and replace them with the new names. For more context on the rename, see
# https://gruntwork-io.slack.com/archives/C0PJF332B/p1611068513028500.

# Bash doesn't have a good way to escape quotes in strings. The official solution is to list multiple strings next
# to each other (https://stackoverflow.com/a/28786747/483528), but that becomes unreadable, especially with regex.
# Therefore, to make our regex more readable, we are declaring clearly-named variables for the special characters we
# want to match, and including those in a string with no need for escaping or crazy concatenation
readonly DOUBLE_QUOTE='"'
readonly SINGLE_QUOTE="'"
readonly BACKTICK='`'
readonly START_OF_LINE='^'
readonly END_OF_LINE='$'
readonly FORWARD_SLASH='\/'
readonly DOT='\.'
readonly WHITESPACE='\s'
readonly OPEN_BRACKET='\['
readonly CLOSE_BRACKET='\]'
readonly OPEN_PAREN='\('
readonly CLOSE_PAREN='\)'
readonly OPEN_CURLY_BRACE='\{'
readonly CLOSE_CURLY_BRACE='\}'

# When replacing old names with new, these are regular expressions for the characters we allow before a name or after.
# We check these characters explicitly to make sure we don't accidentally replace one of the names when it just happens
# to appear as a substring in some unrelated word. E.g., "module-ci" should NOT be replaced in
# "gruntwork-module-circleci-helpers".
readonly ALLOWED_CHARS_BEFORE="($START_OF_LINE|$WHITESPACE|$FORWARD_SLASH|$DOUBLE_QUOTE|$SINGLE_QUOTE|$BACKTICK|$OPEN_BRACKET|$CLOSE_BRACKET|$OPEN_PAREN|$CLOSE_PAREN|$OPEN_CURLY_BRACE|$CLOSE_CURLY_BRACE)"
readonly ALLOWED_CHARS_AFTER="($END_OF_LINE|$WHITESPACE|$FORWARD_SLASH|$DOUBLE_QUOTE|$SINGLE_QUOTE|$BACKTICK|$OPEN_BRACKET|$CLOSE_BRACKET|$OPEN_PAREN|$CLOSE_PAREN|$OPEN_CURLY_BRACE|$CLOSE_CURLY_BRACE|$DOT)"

# The list of repos to replace, in pairs, where the first entry is the old name and the second entry is the second name
# (we use this ugly array instead of a map because the old Bash version on Mac doesn't support maps).
readonly REPLACEMENT_PAIRS=(
  "module-vpc"                  "terraform-aws-vpc"
  "module-aws-monitoring"       "terraform-aws-monitoring"
  "package-directory-services"  "terraform-aws-directory-services"
  "module-file-storage"         "terraform-aws-file-storage"
  "module-ecs"                  "terraform-aws-ecs"
  "module-security"             "terraform-aws-security"
  "cis-compliance-aws"          "terraform-aws-cis-service-catalog"
  "aws-service-catalog"         "terraform-aws-service-catalog"
  "aws-architecture-catalog"    "terraform-aws-architecture-catalog"
  "package-terraform-utilities" "terraform-aws-utilities"
  "module-ci"                   "terraform-aws-ci"
  "module-asg"                  "terraform-aws-asg"
  "module-server"               "terraform-aws-server"
  "package-beanstalk"           "terraform-aws-beanstalk"
  "package-openvpn"             "terraform-aws-openvpn"
  "module-data-storage"         "terraform-aws-data-storage"
  "module-load-balancer"        "terraform-aws-load-balancer"
  "package-zookeeper"           "terraform-aws-zookeeper"
  "package-kafka"               "terraform-aws-kafka"
  "package-messaging"           "terraform-aws-messaging"
  "module-cache"                "terraform-aws-cache"
  "package-static-assets"       "terraform-aws-static-assets"
  "package-elk"                 "terraform-aws-elk"
  "package-mongodb"             "terraform-aws-mongodb"
  "package-lambda"              "terraform-aws-lambda"
  "package-datadog"             "terraform-aws-datadog"
  "package-waf"                 "terraform-aws-waf"
  "package-sam"                 "terraform-aws-sam"
  "module-ci-pipeline-example"  "terraform-aws-ci-pipeline-example"
)

# Finds all files in the repo to replace, taking care to exclude the .git folder, Terraform & Terragrunt scratch
# folders, binary files, and other files we shouldn't be touching. Note that we also skip .go, go.mod, and go.sum
# because our Go code refers to skip versions of our repos, and if we try the new names with the old version numbers,
# Go gives a "module declares its path as: <OLD NAME>" error. So the only way to use the new names with Go code is to
# publish new versions, but to do that, we'd have to build a dependency graph, update repos and publish new versions in
# the right order, update the code to use these new versions, and fix any backwards incompatibilities in these new
# versions, which is far work than we're willing to do for a version number bump right now.
function find_files_to_update {
  find . \
    -not -iwholename '*.git*' \
    -not -iwholename '*.terragrunt-cache*' \
    -not -iwholename '*.terraform*' \
    -not -iwholename '*.png' \
    -not -iwholename '*.jpg' \
    -not -iwholename '*.jpeg' \
    -not -iwholename '*.gif' \
    -not -iwholename '*.bmp' \
    -not -iwholename '*.tiff' \
    -not -iwholename '*.DS_Store*' \
    -not -iwholename '*.go' \
    -not -iwholename '*go.mod' \
    -not -iwholename '*go.sum' \
    -type f
}

# Format a regex replacement string for use with perl. The returned value will be of the format:
#
# s/<REPO_OLD_NAME_1>/<REPO_NEW_NAME_1>/g; s/<REPO_OLD_NAME_2>/<REPO_NEW_NAME_2>/g; s/<REPO_OLD_NAME_3>/<REPO_NEW_NAME_3>/g; ...
#
# This string will allow us to replace multiple values in a single call to Perl.
#
# https://stackoverflow.com/a/8934117/483528
function format_replacement_string {
  local replacements=""
  local i old_name new_name
  for ((i=0;i<"${#REPLACEMENT_PAIRS[@]}";i+=2)); do
    old_name="${REPLACEMENT_PAIRS[i]}"
    new_name="${REPLACEMENT_PAIRS[i+1]}"

    # This is the sed-like regex for the replacements we'll be doing. To help create this regex, I used this handy
    # online regex tester, that not only gives you nice highlighting, but even lets you define a bunch of test cases to
    # check against!
    #
    # https://regexr.com/5l8n4
    #
    replacements+=" s/$ALLOWED_CHARS_BEFORE$old_name$ALLOWED_CHARS_AFTER/\$1$new_name\$2/g;"
  done

  # Strip extra space at start of string: https://stackoverflow.com/a/16623897/483528
  echo "${replacements# }"
}

# The main entrypoint for this script
function update_all_repo_names {
  local replacements
  replacements=$(format_replacement_string)

  local files_to_update
  files_to_update=($(find_files_to_update))

  local file
  for file in "${files_to_update[@]}"; do
    # I originally used sed, but on Mac, sed added an unnecessary newline at the end of every single file it touched,
    # so I switched to Perl. This also has the added benefit of allowing us to process multiple replacements in a
    # single call.
    perl -i -pe "${replacements[@]}" "$file"
  done
}

update_all_repo_names