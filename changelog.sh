#!/bin/bash

git fetch --tags > /dev/null
LATEST_VERSION=$(git tag --list 'v*.*' | sort --version-sort|tail -1)
NEXT_VERSION=$(echo $LATEST_VERSION |cut -d. -f1-2).$[$(echo $LATEST_VERSION |cut -d. -f3)+1]

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
mkdir -p $DIR/changelog/$NEXT_VERSION

echo "Name of changelog file (no yaml extension)"
read fname

fname=$(echo $fname | tr ' ' '_')

FILE=$DIR/changelog/$NEXT_VERSION/$fname.yaml

cat > "$FILE" << EOF
changelog:
  - type: FIX|NON_USER_FACING|BREAKING_CHANGE|DEPENDENCY_BUMP
    description: something useful here
    issueLink: https://github.com/solo-io/gloo/issues/NUMBER
EOF

$EDITOR "$FILE"
git add "$FILE"