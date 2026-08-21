#!/usr/bin/env bash
set -euo pipefail

docker build \
  --platform linux/amd64 \
  --file benzhi.Dockerfile \
  --tag go-sports-league-production-bug001:latest \
  .