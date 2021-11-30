#!/usr/bin/env bash 

# This script will ensure a repo has the necessary pull request and issue templates as
# part of our contributing guide: https://doc.gruntwork.io/guides/contributing/
#
# It creates or replaces the following files:
#   .github/ISSUE_TEMPLATE/bug_report.md
#   .github/ISSUE_TEMPLATE/feature_request.md
#   .github/pull_request_template.md

function create_bug_issue_template {
cat << "EOF" > .github/ISSUE_TEMPLATE/bug_report.md
---
name: Bug report
about: Create a bug report to help us improve.
title: ''
labels: bug
assignees: ''

---

<!--
Have any questions? Check out the contributing docs at https://gruntwork.notion.site/Gruntwork-Coding-Methodology-02fdcd6e4b004e818553684760bf691e,
or ask in this issue and a Gruntwork core maintainer will be happy to help :)
-->

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior including the relevant Terraform/Terragrunt/Packer version number and any code snippets and module inputs you used.

```hcl
// paste code snippets here
```

**Expected behavior**
A clear and concise description of what you expected to happen.

**Nice to have**
- [ ] Terminal output
- [ ] Screenshots

**Additional context**
Add any other context about the problem here.
EOF
}

function create_feature_issue_template {
cat << "EOF" > .github/ISSUE_TEMPLATE/feature_request.md
---
name: Feature request
about: Submit a feature request for this repo.
title: ''
labels: enhancement
assignees: ''

---

<!--
Have any questions? Check out the contributing docs at https://gruntwork.notion.site/Gruntwork-Coding-Methodology-02fdcd6e4b004e818553684760bf691e,
or ask in this issue and a Gruntwork core maintainer will be happy to help :)
-->

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.
EOF
}

function create_pr_template {
cat << "EOF" > .github/pull_request_template.md
<!--
Have any questions? Check out the contributing docs at https://gruntwork.notion.site/Gruntwork-Coding-Methodology-02fdcd6e4b004e818553684760bf691e,
or ask in this Pull Request and a Gruntwork core maintainer will be happy to help :)
Note: Remember to add '[WIP]' to the beginning of the title if this PR is still a work-in-progress. Remove it when it is ready for review!
-->

## Description

<!-- Write a brief description of the changes introduced by this PR -->

### Documentation

<!--
  If this is a feature PR, then where is it documented?

  - If docs exist:
    - Update any references, if relevant.
  - If no docs exist:
    - Create a stub for documentation including bullet points for how to use the feature, code snippets (including from happy path tests), etc.
-->

<!-- Important: Did you make any backward incompatible changes? If yes, then you must write a migration guide! -->

## TODOs

Please ensure all of these TODOs are completed before asking for a review.

- [ ] Ensure the branch is named correctly with the issue number. e.g: `feature/new-vpc-endpoints-955` or `bug/missing-count-param-434`.
- [ ] Update the docs.
- [ ] Keep the changes backward compatible where possible.
- [ ] Run the pre-commit checks successfully.
- [ ] Run the relevant tests successfully.
- [ ] Ensure any 3rd party code adheres with our [license policy](https://www.notion.so/gruntwork/Gruntwork-licenses-and-open-source-usage-policy-f7dece1f780341c7b69c1763f22b1378) or delete this line if its not applicable.


## Related Issues

<!--
  Link to related issues, and issues fixed or partially addressed by this PR.
  e.g. Fixes #1234
  e.g. Addresses #1234
  e.g. Related to #1234
-->
EOF
}

# Ensure the GitHub template directories exist
mkdir -p .github
mkdir -p .github/ISSUE_TEMPLATE

# if the repo does not contain a bug_report.md file, then create one or replace the existing one
if [[ ! -f ".github/ISSUE_TEMPLATE/bug_report.md" ]]; then
    echo "Could not find file at .github/ISSUE_TEMPLATE/bug_report.md, so adding one..."
    create_bug_issue_template
else
    echo "Found file at .github/ISSUE_TEMPLATE/bug_report.md, so replacing it..."
    rm .github/ISSUE_TEMPLATE/bug_report.md
    create_bug_issue_template
fi 

# if the repo does not contain a feature_request.md file, then create one or replace the existing one
if [[ ! -f ".github/ISSUE_TEMPLATE/feature_request.md" ]]; then
    echo "Could not find file at .github/ISSUE_TEMPLATE/feature_request.md, so adding one..."
    create_feature_issue_template
else
    echo "Found file at .github/ISSUE_TEMPLATE/feature_request.md, so replacing it..."
    rm .github/ISSUE_TEMPLATE/feature_request.md
    create_feature_issue_template
fi 

# if the repo does not contain a pull_request_template.md file, then create one or replace the existing one
if [[ ! -f ".github/pull_request_template.md" ]]; then
    echo "Could not find file at .github/pull_request_template.md, so adding one..."
    create_pr_template
else
    echo "Found file at .github/pull_request_template.md, so replacing it..."
    rm .github/pull_request_template.md
    create_pr_template
fi 
