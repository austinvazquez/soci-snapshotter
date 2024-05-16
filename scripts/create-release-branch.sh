#!/usr/bin/env bash

#   Copyright The Soci Snapshotter Authors.

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

# A script to create a release branch on origin from the given commit.
#
# Usage: bash create-release-branch.sh [-b|--base] [-l|--live] [-r|--rollback] <MAJOR_MINOR_VERSION>

set -eux -o pipefail

BASE=""
DRYRUN=false
ROLLBACK=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --base|-b)
      shift # past argument
      BASE=$1
      shift # past value
      ;;
    --dry-run|-d)
      DRYRUN=true
      shift # past argument
      ;;
    --rollback|-r)
      ROLLBACK=true
      shift # past argument
      ;;
    --*|-*)
      echo "Unknown option $1"
      exit 1
      ;;
    *)
      VERSION=$1
      shift # past argument
      ;;
  esac
done

sanitize_input() {
  # Strip 'v' prefix from input if present.
  VERSION=${VERSION/v/}
  [[ $VERSION =~ ^[0-9]+\.[0-9]+$ ]] || (echo "Error: version does not match expected <major>.<minor> format" && exit 1)

  if [ -n "$BASE" ]; then
    [[ $BASE =~ ^[0-9a-fA-F]{7,40}$ ]] || (echo "Error: base commit does not match expected short|full format" && exit 1)
    FOUND=$(git log --pretty=format:"%H" | grep "$BASE")
    [ -n "$FOUND" ] || (echo "Error: base commit not found in history" && exit 1)
  fi
}

sanitize_input

PUSH_OPTS=()
if [ $DRYRUN = true ]; then
  echo "Dry-run: setting '--dry-run' for git push"
  PUSH_OPTS+=("--dry-run")
fi

if [ $ROLLBACK = true ]; then
  echo "Rollback: setting '--delete' for git push"
  PUSH_OPTS+=("--delete")
else
  git checkout -b "release/${VERSION}" "${BASE}"
fi

git push "${PUSH_OPTS[@]}" origin "release/${VERSION}"
