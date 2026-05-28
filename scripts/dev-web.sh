#!/usr/bin/env bash
set -euo pipefail

npm --prefix web run dev -- --host 0.0.0.0
