#!/bin/bash

set -ex

protoc --version

if [ ! -f .gitignore ]; then
  echo "_output" > .gitignore
fi


git init
git config user.email "you@example.com"
git config --global user.name "Your Name"
git add .
git commit -m "set up dummy repo for diffing" -q


PATH=/workspace/gopath/bin:$PATH

set +e

make generated-code -B
if [[ $? -ne 0 ]]; then
  echo "Code generation failed"
  exit 1;
fi
if [[ $(git status --porcelain | wc -l) -ne 0 ]]; then
  echo "Generating code produced a non-empty diff."
  echo "Try running 'dep ensure && make install-codegen-deps generated-code -B' then re-pushing."
  git status --porcelain
  git diff | cat
  exit 1;
fi
