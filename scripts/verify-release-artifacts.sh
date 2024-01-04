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

# A script to verify artifacts from release automation.

function usage {
    echo "Usage: $0 <release>"
    exit 1
}

if [ $# -eq 0 ]; then
    echo "$0: Missing required argument"
    usage
fi

release=$1

tarballs=("soci-snapshotter-${release}-linux-amd64.tar.gz" "soci-snapshotter-${release}-linux-amd64-static.tar.gz")
expected_contents=("soci-snapshotter-grpc" "soci" "THIRD_PARTY_LICENSES" "NOTICE.md")
release_is_valid=true

for t in ${tarballs[@]}; do
    # Verify each expected tarball was generated.
    if [[ ! -e $t ]]; then
        echo "Missing $t"
        release_is_valid=false
        continue
    fi

    # Verify the tarball's checksum is present and valid.
    if [[ ! -e "$t.sha256sum" ]] || ( ! sha256sum -c $t.sha256sum 2>/dev/null); then
        echo "Checksum for $t is missing or invalid"
        release_is_valid=false
        continue
    fi

    # Verify the tarball contains the expected contents.
    found_contents=$(tar -tvf $t | awk '{print $6}')
    for file in ${found_contents[@]}; do
        if [[ ! ${expected_contents[@]} =~ $file ]]; then
            echo "Unexpected file $file in $t"
            release_is_valid=false
        fi
    done
done

if ( ! ${release_is_valid} ); then
    echo "Release is invalid"
    exit 1
fi

echo "Release is valid"
exit 0
