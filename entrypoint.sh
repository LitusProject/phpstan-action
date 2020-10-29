#! /bin/sh

cd "$GITHUB_WORKSPACE"

phpstan analyze \
    --error-format=json \
    --no-interaction \
    --no-progress \
    | phpstan-action
