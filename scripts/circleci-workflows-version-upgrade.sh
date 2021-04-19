#!/usr/bin/env bash

echo "Upgrading CircleCI workflows syntax to 2..."

 yq w -i .circleci/config.yml 'workflows.version' 2

 # Remove stray merge tags that yq adds to the final output
 sed -i '/!!merge /d' .circleci/config.yml
