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

# A script to assert version in the getting started guide was updated
# correctly by GitHub Actions workflow.
#
# Usage: bash assert-getting-started-guide-version-diff.sh <RELEASE_TAG>

set -eux -o pipefail

release_tag=$1

# Strip 'v' prefix from tag if not already stripped.
release_version=${release_tag/v/}

# Disable warning for A && B || C is not if-then-else; C may run when A is true.
# Branch B contains exit, so C will not run when A and B branches fail.
# This is intended to have the assertion fail if the diff is empty.
# shellcheck disable=SC2015
diff_output=$(git diff --exit-code) && {
  echo "Error: no changes made; expected getting started version to be updated to \"${release_version}\"" && exit 1
} || {
  echo "${diff_output}"

  if [[ "${diff_output}" == *"+version=\"${release_version}\""* ]]; then
    echo "Diff looks good!"
  else
    echo "Error: release version not set properly" && exit 1
  fi
}
