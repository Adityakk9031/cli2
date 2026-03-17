#!/usr/bin/env bash

set -euo pipefail

: "${RUN_URL:?RUN_URL is required}"
: "${E2E_AGENT:?E2E_AGENT is required}"
: "${TRIAGE_OUTPUT_FILE:?TRIAGE_OUTPUT_FILE is required}"

mkdir -p "$(dirname "$TRIAGE_OUTPUT_FILE")"

claude --plugin-dir .claude/plugins/e2e \
  -p "/e2e:triage-ci ${RUN_URL} --agent ${E2E_AGENT}" \
  2>&1 | tee "$TRIAGE_OUTPUT_FILE"
