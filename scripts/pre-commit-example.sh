#!/usr/bin/env bash 

function add_precommit {
cat << EOF > .pre-commit-config.yaml
repos:
  - repo: https://github.com/gruntwork-io/pre-commit
    rev: <VERSION> # Get the latest from: https://github.com/gruntwork-io/pre-commit/releases
    hooks:
      - id: terraform-fmt
      - id: terraform-validate
      - id: tflint
      - id: shellcheck
      - id: gofmt
      - id: golint
EOF
}

echo "Running pre-commit example.sh..."

if [[ ! -f .pre-commit-config.yaml ]]; then 
	echo ".pre-commit-conifg.yaml file does not already exist. Adding it now..."
	add_precommit
else 
	echo "Found existing .pre-commit-config.yaml file. Nothing to do."

fi 


