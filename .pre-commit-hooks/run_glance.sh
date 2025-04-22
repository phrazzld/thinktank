#!/bin/bash
# Post-commit hook to run glance on the repository
# This generates or updates glance.md files that provide directory overviews

set -e

echo "Starting glance to update directory overviews in the background..."
(
  glance ./
  echo "Glance directory overviews updated successfully."
) > /tmp/glance_output.log 2>&1 &
