#!/usr/bin/env bash
set -u

DIVERGED=$(git merge-base --fork-point origin/master > /dev/null; echo $?)

if [[ $DIVERGED == 0 ]]; then
  if [[ $BUILDKITE_TAG == "" ]]; then
    if [[ $BUILDKITE_BRANCH == "master" ]]; then
      CI_BYPASS=$(git diff --name-only HEAD~1 | sed -rn '/^(BREAKING.md|CONTRIBUTING.md|README.md|docs\/.*)/!{q1}' && echo true || echo false)
    else
      CI_BYPASS=$(git diff --name-only `git merge-base --fork-point origin/master` | sed -rn '/^(BREAKING.md|CONTRIBUTING.md|README.md|docs\/.*)/!{q1}' && echo true || echo false)
    fi
  else
    CI_BYPASS="false"
  fi
else
  CI_BYPASS="false"
fi

cat << EOF
env:
  CI_BYPASS: ${CI_BYPASS}

steps:
  - label: ":docker: Image Deployments"
    command: ".buildkite/steps/deployimages.sh | buildkite-agent pipeline upload"
    if: build.env("CI_BYPASS") != "true"

  - wait:
    if: build.env("CI_BYPASS") != "true"

  - label: ":docker: Deploy Manifests"
    command: "authelia-scripts docker push-manifest"
    env:
      DOCKER_CLI_EXPERIMENTAL: "enabled"
    if: build.env("CI_BYPASS") != "true"

  - label: ":github: Deploy Artifacts"
    command: "ghartifacts.sh"
    depends_on:
      - "build-docker-linux-amd64"
      - "build-docker-linux-arm32v7"
      - "build-docker-linux-arm64v8"
    retry:
      automatic: true
    agents:
      upload: "fast"
    key: "artifacts"
    if: build.tag != null

  - label: ":linux: Deploy AUR"
    command: ".buildkite/steps/aurpackages.sh | buildkite-agent pipeline upload"
    depends_on: ~
    if: build.tag != null || build.branch == "master" && build.env("CI_BYPASS") != "true"

  - label: ":book: Deploy Documentation"
    command: "syncdoc.sh"
    depends_on: ~
    agents:
      upload: "fast"
    if: build.branch == "master"
EOF