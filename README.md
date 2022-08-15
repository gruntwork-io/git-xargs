[![Go Report Card](https://goreportcard.com/badge/github.com/gruntwork-io/git-xargs)](https://goreportcard.com/report/github.com/gruntwork-io/git-xargs)
[![gruntwork-io](https://circleci.com/gh/gruntwork-io/git-xargs.svg?style=svg)](https://circleci.com/gh/gruntwork-io/git-xargs)
[![Homebrew](https://img.shields.io/badge/dynamic/json.svg?url=https://formulae.brew.sh/api/formula/git-xargs.json&query=$.versions.stable&label=homebrew)](https://formulae.brew.sh/formula/git-xargs)

# Table of contents

- [Introduction](#introduction)
- [Reference](#reference)
- [Contributing](#contributing)

# Introduction

## Overview

![git-xargs CLI](./docs/git-xargs-banner.png)

`git-xargs` is a command-line tool (CLI) for making updates across multiple GitHub repositories with a single command. You give `git-xargs`:

1. a script or a command to run
2. a list of repos

and `git-xargs` will:

1. clone each repo
1. run your specified script or command against it
1. commit any changes
1. open pull requests
1. provide a detailed report of everything that happened

Git-xargs leverages goroutines to perform the repo-updating work in parallel, so it is very fast.

For example, have you ever needed to add a particular file across many repos at once? Or to run a search and replace to change your company or product name across 150 repos with one command? What about upgrading Terraform modules to all use the latest syntax? How about adding a CI/CD configuration file, if it doesn't already exist, or modifying it in place if it does, but only on a subset of repositories you select?
You can handle these use cases and many more with a single `git-xargs` command.

## Example: writing a new file to every repo in your GitHub organization

As an example, let's use `git-xargs` to create a new file in every repo:

```
git-xargs \
  --branch-name test-branch \
  --github-org <your-github-org> \
  --commit-message "Create hello-world.txt" \
  touch hello-world.txt
```

Here's what it looks like in action:

![git-xargs to the rescue!](docs/git-xargs-demo.gif)

In this example, every repo in your org will have a new file named hello-world.txt written to it with the contents "Hello, World!". You'll then receive an easy-to-read printout of exactly what happened on `STDOUT`:

```
*****************************************************************
  GIT-XARGS RUN SUMMARY @ 2021-04-12 23:05:18.478435534 +0000 UTC
  Runtime in seconds: 4
*****************************************************************


COMMAND SUPPLIED

[touch hello-world.txt]

 REPOS SUPPLIED VIA --repos FILE FLAG
│────────────────────────│────────────────────────│
│ ORGANIZATION NAME (5)  │ URL                    │
│────────────────────────│────────────────────────│
│ zack-test-org          │ terraform-aws-asg      │
│ zack-test-org          │ terraform-aws-vpc      │
│ zack-test-org          │ terraform-aws-security │
│ zack-test-org          │ terraform-aws-eks      │
│ zack-test-org          │ circleci-test-1        │
│────────────────────────│────────────────────────│

 ALL REPOS THAT WERE TARGETED FOR PROCESSING AFTER FILTERING MISSING / MALFORMED REPOS
│───────────────────│────────────────────────────────────────────────────│
│ REPO NAME         │ REPO URL                                           │
│───────────────────│────────────────────────────────────────────────────│
│ terraform-aws-vpc │ https://github.com/zack-test-org/terraform-aws-vpc │
│ terraform-aws-eks │ https://github.com/zack-test-org/terraform-aws-eks │
│ circleci-test-1   │ https://github.com/zack-test-org/circleci-test-1   │
│───────────────────│────────────────────────────────────────────────────│


 REPOS THAT WERE SUCCESSFULLY CLONED TO THE LOCAL FILESYSTEM
│───────────────────│────────────────────────────────────────────────────│
│ REPO NAME         │ REPO URL                                           │
│───────────────────│────────────────────────────────────────────────────│
│ terraform-aws-eks │ https://github.com/zack-test-org/terraform-aws-eks │
│ circleci-test-1   │ https://github.com/zack-test-org/circleci-test-1   │
│ terraform-aws-vpc │ https://github.com/zack-test-org/terraform-aws-vpc │
│───────────────────│────────────────────────────────────────────────────│


 REPOS THAT SHOWED FILE CHANGES TO THEIR WORKING DIRECTORY FOLLOWING COMMAND EXECUTION
│───────────────────│────────────────────────────────────────────────────│
│ REPO NAME         │ REPO URL                                           │
│───────────────────│────────────────────────────────────────────────────│
│ terraform-aws-eks │ https://github.com/zack-test-org/terraform-aws-eks │
│ terraform-aws-vpc │ https://github.com/zack-test-org/terraform-aws-vpc │
│ circleci-test-1   │ https://github.com/zack-test-org/circleci-test-1   │
│───────────────────│────────────────────────────────────────────────────│


 REPOS THAT WERE SUPPLIED BY USER BUT DON'T EXIST (404'D) VIA GITHUB API
│────────────────────────│──────────│
│ REPO NAME              │ REPO URL │
│────────────────────────│──────────│
│ terraform-aws-asg      │          │
│ terraform-aws-security │          │
│────────────────────────│──────────│


 REPOS WHOSE SPECIFIED BRANCHES DID NOT EXIST ON THE REMOTE, AND SO WERE FIRST CREATED LOCALLY
│───────────────────│────────────────────────────────────────────────────│
│ REPO NAME         │ REPO URL                                           │
│───────────────────│────────────────────────────────────────────────────│
│ terraform-aws-eks │ https://github.com/zack-test-org/terraform-aws-eks │
│ terraform-aws-vpc │ https://github.com/zack-test-org/terraform-aws-vpc │
│ circleci-test-1   │ https://github.com/zack-test-org/circleci-test-1   │
│───────────────────│────────────────────────────────────────────────────│


*****************************************************
  PULL REQUESTS OPENED
*****************************************************
│───────────────────│────────────────────────────────────────────────────────────│
│ REPO NAME         │ PR URL                                                     │
│───────────────────│────────────────────────────────────────────────────────────│
│ circleci-test-1   │ https://github.com/zack-test-org/circleci-test-1/pull/82   │
│ terraform-aws-eks │ https://github.com/zack-test-org/terraform-aws-eks/pull/81 │
│ terraform-aws-vpc │ https://github.com/zack-test-org/terraform-aws-vpc/pull/77 │
│───────────────────│────────────────────────────────────────────────────────────│

```

## Getting started

### Installation option 1: Homebrew

If you are [Homebrew](https://brew.sh/) user, you can install by running

```bash
$ brew install git-xargs
```

### Installation option 2: Installing published binaries

1. **Download the correct binary for your platform**. Visit [the releases
   page](https://github.com/gruntwork-io/git-xargs/releases) and download the correct binary depending on your system.
   Save it to somewhere on your `PATH`, such as `/usr/local/bin/git-xargs`.

1. **Set execute permissions**. For example, on Linux or Mac, you'd run:

      ```bash
      chmod u+x /usr/local/bin/git-xargs
      ```

1. **Check it's working**. Run the version command to ensure everything is working properly:

      ```bash
      git-xargs --version
      ```

### Installation option 3: Run go install or go get

1. **Ensure you have Golang installed and working properly on your system.** [Follow the official Golang install guide](https://golang.org/doc/install) to get started.

1. **Run go install to install the latest release of git-xargs**:
     ```bash
     go install github.com/gruntwork-io/git-xargs@latest
     ```

1. **Alternatively, use go install to select a specific release of git-xargs**:
     ```bash
     go install github.com/gruntwork-io/git-xargs@v0.0.5
     ```

1. **If you have Go 1.16 or earlier, you can use get**
     ```bash
     go get github.com/gruntwork-io/git-xargs
     go get github.com/gruntwork-io/git-xargs@v0.0.5
     ```

### Try it out!

1. **Export a valid GitHub token**. See the guide on [Github personal access
   tokens](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token)
   for information on how to generate one. For example, on Linux or Mac, you'd run:

      ```bash
      export GITHUB_OAUTH_TOKEN=<your-secret-github-oauth-token>
      ```

1. **Provide a script or command and target some repos**. Here's a simple example of running the `touch` command in
   every repo in your GitHub organization. Follow the same pattern to start running your own scripts and commands
   against your own repos!

      ```bash
      git-xargs \
        --branch-name "test-branch" \
        --commit-message "Testing git-xargs" \
        --github-org <enter-your-github-org-name> \
        touch git-xargs-is-awesome.txt
      ```

# Reference

## How to supply commands or scripts to run

The API for `git-xargs` is:

```
git-xargs [-flags] <CMD>
```

Where `CMD` is either the full path to a (Bash, Python, Ruby, etc) script on your local system or a binary. Note that, because the tool supports Bash scripts, Ruby scripts, Python scripts, etc, you must include the full filename for any given script, including its file extension.

In other words, all the following usages are valid:

```
git-xargs --repo gruntwork-io/cloud-nuke \
   --repo gruntwork-io/terraform-aws-eks \
   --branch-name my-branch \
   /usr/local/bin/my-bash-script.sh
```

```
git-xargs --repos ./my-repos.txt \
  --branch-name my-other-branch \
  touch file1.txt file2.txt
```

```
git-xargs --github-org my-github-org \
  --branch-name my-new-branch \
  "$(pwd)/scripts/my-ruby-script.rb"
```

## Debugging runtime errors

By default, `git-xargs` will conceal runtime errors as they occur because its log level setting is `INFO` if not overridden by the `--loglevel` flag.

To see all errors your script or command may be generating, be sure to pass `--loglevel DEBUG` when running your `git-xargs` command, like so:

```
git-xargs --loglevel DEBUG \
	--repo zack-test-org/terraform-aws-eks \
	--branch-name master \
	--commit-message "add blank file" \
	--skip-pull-requests touch foo.txt
```

When the log level is set to `debug` you should see new error output similar to the following:

```
Total 195 (delta 159), reused 27 (delta 11), pack-reused 17  Repo=terraform-aws-eks
[git-xargs] DEBU[2021-06-29T12:11:31-04:00] Created branch                                Branc
h Name=refs/heads/master Repo=terraform-aws-eks
[git-xargs] DEBU[2021-06-29T12:11:31-04:00] Error creating new branch                     Error
="a branch named \"refs/heads/master\" already exists" Repo=terraform-aws-eks
[git-xargs] DEBU[2021-06-29T12:11:31-04:00] Error encountered while processing repo       Error
="a branch named \"refs/heads/master\" already exists" Repo name=terraform-aws-eks

```

## Rate Limiting

git-xargs attempts to be a good citizen as regards consumption of the GitHub API. git-xargs conforms to GitHub's API [integration guidelines](https://docs.github.com/en/rest/guides/best-practices-for-integrators#dealing-with-secondary-rate-limits).

In addition, git-xargs includes several features and flags to help you: 

1. run jobs without tripping GitHub's rate limits
1. recover when rate limited by automatically re-trying failed pull requests again, all while honoring the GitHub rate limits

**Distinct processing channels for expensive and non-expensive work**

git-xargs distinguishes between work that is safe to perform in parallel, such as certain git operations, and work that must be done with consideration of resource constraints, such as issuing open pull requests to GitHub's API. Therefore, git-xargs is able to perform all concurrency-safe work as quickly as possible by leveraging goroutines, while treating the more expensive open pull request API calls separately. 

Pull requests are handled on a separate channel so that they can be buffered and retried in accordance with rate limiting feedback git-xargs is receiving from GitHub's API.  

This means that git-xargs performs all the work upfront that it can as quickly as possible, and then moves on to serially process the pull request jobs that have resulted from the concurrency-safe work of cloning repositories, making file changes, checking the git worktree, etc.

**Automatic pull request retries when rate limited**

By default, git-xargs will re-attempt opening a pull request that failed due to rate limiting. The `--max-pull-request-retries` flag allows you to specify how many times you'd like a given pull request to be re-attempted in case of failure due to rate limiting. By default, the value is `3`, meaning that if you do not pass this flag, `git-xargs` will retry all rate-limit-blocked pull requests 3 times.

**Automatic backoff when rate limiting is detected**

When git-xargs detects that it has been rate limited by GitHub, it begins requeuing failed pull requests for retry, but with an additional, larger buffer of time in between the next attempt. This larger buffer of time is intended to allow GitHub rate limit status to return to baseline for the git-xargs client. The `--seconds-to-wait-when-ratelimited` flag specifies the number of seconds to wait when git-xargs detects it has been rate limited. Note that this extra buffer of time is in addition to the value specified by the `--seconds-between-prs` flag.

When they are provided by GitHub, git-xargs will instead use the values of any `Retry-After` header, or the delta of seconds between the current time and the rate limit error's reset time. When these two values are not available, git-xargs falls back to the user-specified value of `--seconds-to-wait-when-ratelimited`, if the flag was passed, or the default value, which is 60 seconds.

**Specifying the pause between pull requests**

The GitHub's API [integration guidelines](https://docs.github.com/en/rest/guides/best-practices-for-integrators#dealing-with-secondary-rate-limits) specify that clients should pause at least 1 second in between consecutive requests against expensive endpoints, such as the one for opening a new pull request. As a result, git-xargs defaults to pausing 1 second in between each pull request that is opened. The flag `--seconds-between-prs` allows you to modify this value. For example, if you were to pass `--seconds-between-prs 30`, then git-xargs will sleep for half a minute between issuing pull request API calls to GitHub.

**Guidelines and observations from testing**

In local testing, the actual thresholds for when GitHub's secondary rate limits kick in can vary depending on the number of repos you're targeting and how long your script or command takes to complete. Secondary rate limits have been observed on jobs targeting as few as 10 repositories. If you are using the rate limiting flags with reasonable values, your job should never be rate-limited in the first place. 

If your job is consistently being rate-limited, try incrementally increasing the value you pass with the  `--seconds-between-prs` flag. Passing a higher value will increase the overall time your job takes to complete, but it will also greatly decrease the likelihood of your job tripping GitHub's rate limits at all. Passing a lower value, or not passing the flag at all, will greatly increase the likelihood that your job is rate limited by GitHub. 


## Branch behavior

Passing the `--branch-name` (`-b`) flag is required when running `git-xargs`. If you specify the name of a branch that exists on your remote, its latest changes will be pulled locally prior to your command or script being run. If you specify the name of a new branch that does not yet exist on your remote, it will be created locally and pushed once your changes are committed.

## Default repository branch

Any pull requests opened will be opened against the repository's default branch (whether that's `main`, or `master` or something else). You can supply an additional `--base-branch-name` flag to change the target for your pull requests. Be aware that this will override the base branch name for **ALL** targeted repositories.

## Git file staging behavior

Currently, `git-xargs` will find and add any and all new files, as well as any existing files that were modified, within your repo and stage them prior to committing. If your script or command creates a new file, it will be committed. If your script or command edits an existing file, that change will also be committed.

## Paths and script locations

Scripts may be placed anywhere on your system, but you are responsible for providing absolute paths to your scripts when invoking `git-xargs`:

```
git-xargs \
  --branch-name upgrade-tf-14 \
  --commit-message "Update modules to Terraform 0.14" \
  --repos data/batch3.txt \
  $(pwd)/scripts/my-ruby-script.rb
```

or

```
git-xargs \
  --branch-name upgrade-tf-14 \
  --commit-message "Update modules to Terraform 0.14" \
  --repos data/batch3.txt \
  /usr/local/bin/my-ruby-script.rb
```

If you need to compose more complex behavior into a single pull request, write a wrapper script that executes all your commands, or place all your logic into one script.

## How to target repos to run your scripts against

`git-xargs` supports **four** methods of targeting repos to run your selected scripts against. They are processed in
the order listed below, with whichever option is found first being used, and all others after it being ignored.

### Option #1: GitHub organization lookup

If you want the tool to find and select every repo in your GitHub organization, you can pass the name of your organization via the `--github-org` flag:

```
git-xargs \
  --commit-message "Update copyright year" \
  --github-org <your-github-org> \
  "$(pwd)/scripts/update-copyright-year.sh"
```

This will signal the tool to look up, and page through, every repository in your GitHub organization and execute the scripts you passed.

### Option #2: Flat file of repository names

Oftentimes, you want finer-grained control over the exact repos you are going to run your script against. In this case, you can use the `--repos` flag and supply the path to a file defining the exact repos you want the tool to run your selected scripts against, like so:

```
git-xargs \
  --commit-message "Update copyright year" \
  --repos data/batch2.txt \
  "$(pwd)/scripts/update-copyright-year.sh"
```

In this example, batch2.txt looks like this:

```
gruntwork-io/infrastructure-as-code-training
gruntwork-io/infrastructure-live-acme
gruntwork-io/infrastructure-live-multi-account-acme
gruntwork-io/infrastructure-modules-acme
gruntwork-io/infrastructure-modules-multi-account-acme
```

Flat files contain one repo per line, each repository in the format of `<github-organization>/<repo-name>`. Commas, trailing or preceding spaces, and quotes are all filtered out at runtime. This is done in case you end up copying your repo list from a JSON list or CSV file.

### Option #3: Pass in repos via command line args

Another way to get fine-grained control is to pass in the individual repos you want to use via one or more `--repo`
arguments:

```
git-xargs \
  --commit-message "Update copyright year" \
  --repo gruntwork-io/terragrunt \
  --repo gruntwork-io/terratest \
  --repo gruntwork-io/cloud-nuke \
  "$(pwd)/scripts/update-copyright-year.sh"
```

### Option #4: Pass in repos via stdin

And one more (Unix-philosophy friendly) way to get fine-grained control is to pass in the individual repos you want to
use by piping them in via `stdin`, separating repo names with whitespace or newlines:

```
echo "gruntwork-io/terragrunt gruntwork-io/terratest" | git-xargs \
  --commit-message "Update copyright year" \
  "$(pwd)/scripts/update-copyright-year.sh"
```

## Notable flags

`git-xargs` exposes several flags that allow you to customize its behavior to better suit your needs. For the latest info on flags, you should run `git-xargs --help`. However, a couple of the flags are worth explaining more in depth here:

| Flag                                  | Description                                                                                                                                                                                                                                                                                                                                                                                                                                     | Type    | Required |
| ------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | -------- |
| `--branch-name`                       | You must specify the name of the branch to make your local and remote changes on. You can further control branching behavior via `--skip-pull-requests` as explained below.                                                                                                                                                                                                                                                                     | String  | Yes      |
| `--loglevel`                          | Specify the log level of messages git-xargs should print to STDOUT at runtime. By default, this is INFO - so only INFO level messages will be visible. Pass DEBUG to see runtime errors encountered by your scripts or commands. Accepted levels are TRACE, DEBUG, INFO, WARNING, ERROR, FATAL and PANIC. Default: `INFO`.                                                                                                                      | String  | No       |
| `--repos`                             | If you want to specify many repos and manage them in files (which makes batching and testing easier) then use this flag to pass the filepath to a repos file. See [the repos file format](#option-2-flat-file-of-repository-names) for more information.                                                                                                                                                                                        | String  | No       |
| `--repo`                              | Use this flag to specify a single repo, e.g., `--repo gruntwork-io/cloud-nuke`. Can be passed multiple times to target several repos.                                                                                                                                                                                                                                                                                                           | String  | No       |
| `--github-org`                        | If you want to target every repo in a Github org that your GITHUB_OAUTH_TOKEN has access to, pass the name of the Organization with this flag, to page through every repo via the Github API and target it.                                                                                                                                                                                                                                     | String  | No       |
| `--commit-message`                    | The commit message to use when creating commits. If you supply this flag, but neither the optional `--pull-request-title` or `--pull-request-description` flags, then the commit message value will be used for all three. Default: `[skip ci] git-xargs programmatic commit`. Note that, by default, git-xargs will prepend \"[skip ci]\" to commit messages unless you pass the `--no-skip-ci` flag. If you wish to use an alternative prefix other than [skip ci], you can add the literal string to your --commit-message value.                                                                                                                                                                            | String  | No       |
| `--skip-pull-requests`                | If you don't want any pull requests opened, but would rather have your changes committed directly to your specified branch, pass this flag. Note that it won't work if your Github repo is configured with branch protections on the branch you're trying to commit directly to! Default: `false`.                                                                                                                                              | Boolean | No       |
| `--skip-archived-repos`               | If you want to exclude archived (read-only) repositories from the list of targeted repos, pass this flag. Default: `false`.                                                                                                                                                                                                                                                                                                                     | Boolean | No       |
| `--dry-run`                           | If you are in the process of testing out `git-xargs` or your initial set of targeted repos, but you don't want to make any changes via the Github API (pushing your local changes or opening pull requests) you can pass the dry-run flag. This is useful because the output report will still tell you which repos would have been affected, without actually making changes via the Github API to your remote repositories. Default: `false`. | Boolean | No       |
| `--draft`                             | Whether to open pull requests in draft mode. Draft pull requests are available for public GitHub repositories and private repositories in GitHub tiered accounts. See [Draft Pull Requests](https://docs.github.com/en/github/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests#draft-pull-requests) for more details. Default: false.                                                   | Boolean | No       |
| `--seconds-between-prs`               | The number of seconds to wait between opening serial pull requests. If you are being rate limited, continue to increase this value until rate limiting eases. Note, this value cannot be negative, so if you pass a value less than 1, the seconds to wait between pull requests will be set to 1 second. Default: `1` second.                                                                                                                 | Integer | No       |
| `--max-pr-retries`                    | The number of seconds to wait between opening serial pull requests. If you are being rate limited, continue to increase this value until rate limiting eases. Default: `3` seconds.                                                                                                                                                                                                                                                            | Integer | No       |
| `--seconds-to-wait-when-rate-limited` | The number of seconds to pause once git-xargs has detected it has been rate limited. Note that this buffer is in addition to the value of --seconds-between-prs. If you are regularly being rate limited, increase this value until rate limiting eases. Default: `60` seconds.                                                                                                                                                                | Integer | No       |
| `--no-skip-ci` | By default, git-xargs will prepend \"[skip ci]\" to its commit messages to prevent large git-xargs jobs from creating expensive CI jobs excessively. If you pass the `--no-skip-ci` flag, then git-xargs will not prepend \"[skip ci]\". Default: false, meaning that \"[skip ci]\" will be prepended to commit messages.                                                                                                                                                                | Bool | No       |

## Best practices, tips and tricks

### Write your script to run against a single repo

Write your script as if it's operating on a single repo, then target many repos with `git-xargs`. Remember that at runtime, each of the scripts you select will be run, in the order you specify, once per repo that you've targeted.

### Handling prerequisites and third party binaries

It is currently assumed that bash script authors will be responsible for checking for prerequisites within their own scripts. If you are adding a new bash script to accomplish some new task across repos, consider using the [Gruntwork bash-commons assert_is_installed pattern](https://github.com/gruntwork-io/bash-commons/blob/3cb3c7160fb72b7411af184300bf077caede37e4/modules/bash-commons/src/assert.sh#L15) to ensure the operator has any required binaries installed.

### Grouping your repos into separate batches

This is a pattern that ended up working out well for us as we wrote and executed more and more ambitious scripts across our many repos as a team:
By breaking your target repos into separate batches, (batch1.txt, batch2.txt, batch3.txt) and starting with a few repos (or even one repo!) in the initial batches, and then gradually expanding the batches in size, you can easily test your new scripts against a few repos and double check the generated pull requests for any issues prior to widening your target batches.

## How git-xargs works

This section provides a more in-depth look at how the `git-xargs` tool works under the hood.

1. git-xargs will clone each of your selected repos to your machine to the `/tmp/` directory of your local machine. The name of each repo, plus a random number each run, are concatenated together to form the local clone name to make the local repo easier to find in case you need to debug your script locally, e.g., `terraform-aws-module-security3978298`.
1. it will checkout a local branch (whose name you must specify with the `--branch-name` flag)
1. it will run all your selected scripts against your selected repos
1. it will commit any changes in each of the repos (with a commit message you can optionally specify via the `--commit-message` flag)
1. it will push your local branch with your new commits to your repo's remote
1. it will call the GitHub API to open a pull request with a title and description that you can optionally specify via the `--pull-request-title` and `--pull-request-description` flags, respectively, unless you pass the `--skip-pull-requests` flag
1. it will print out a detailed run summary to STDOUT that explains exactly what happened with each repo and provide links to successfully opened pull requests that you can quickly follow from your terminal. If any repos encountered errors at runtime (whether they weren't able to be cloned, or script errors were encountered during processing, etc) all of this will be spelled out in detail in the final report, so you know exactly what succeeded and what went wrong.

## Tasks this tool is well-suited for

The following is a non-exhaustive list of potential use cases for `git-xargs`:

- Add a LICENSE file to all of your GitHub repos, interpolating the correct year and company name into the file
- For every existing LICENSE file across all your repos, update their copyright date to the current year
- Update the CI build configuration in all of your repos by modifying the `.circleci/config.yml` file in each repo using a tool such as `yq`
- Run `sed` commands to update or replace information across README files
- Add new files to repos
- Delete specific files, when present, from repos
- Modify `package.json` files in-place across repos to bump a node.js dependency using `jq` https://stedolan.github.io/jq/
- Update your Terraform module library from Terraform 0.13 to 0.14.
- Remove stray files of any kind, when found, across repos using `find` and its `exec` option
- Add baseline tests to repos that are missing them by copying over a common local folder where they are defined
- Refactor multiple Golang tools to use new libraries by executing `go get` to install and uninstall packages, and modify the source code files' import references

_If you can instrument the logic in a script, you can use `git-xargs` to run it across your repos!_

# Contributing

## Contributing scripts to this project

We hope that this tool will help save you some time as you apply it to your own automations and maintenance tasks. We also welcome the community to contribute back scripts that everyone can use and benefit from.

Initially, we'll add these scripts to the `./scripts` directory in this repository and will eventually organize them into sub-folders depending on their purposes / use cases. If you would like to have your script considered for inclusion in this repo, please first ensure that it is:

- **High quality**: meaning free of typos, any obvious bugs or security issues
- **Generic**: meaning that it is likely to be of general use to many different people and organizations, and free of any proprietary tooling or secrets

Once you've done this, please feel free to open a pull request adding your script to the `./scripts` directory for consideration.

Thanks for contributing back! Our hope is that eventually this repo will contain many useful generic scripts for common maintenance and upgrading tasks that everyone can leverage to save time.

## Building the binary from source

Clone this repository and then run the following command from the root of the repository:

```
go build
```

The `git-xargs` binary will be present in the repository root.

## Running the tool without building the binary

Alternatively, you can run the tool directly without building the binary, like so:

```
./go run main.go \
  --branch-name test-branch \
  --commit-message "Add MIT License" \
  --repos data/test-repos.txt \
  $(pwd)/scripts/add-license.sh
```

This is especially helpful if you are developing against the tool and want to quickly verify your changes.

## Running tests

Tests are included within their respective packages.

```
go test -v ./...
```

## License

This code is released under the Apache 2.0 License. See [LICENSE.txt](/LICENSE.txt)
