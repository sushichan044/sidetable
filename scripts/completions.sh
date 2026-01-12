#!/bin/sh
set -e

rm -rf completions
mkdir completions
for sh in bash zsh fish; do
    go run ./cmd/sidetable completion "$sh" >"completions/sidetable.$sh"
done
