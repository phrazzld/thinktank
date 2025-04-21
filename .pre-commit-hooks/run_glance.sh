#!/bin/bash
# Post-commit hook to run glance on the repository
# This generates or updates glance.md files that provide directory overviews

set -e

echo "Running glance to update directory overviews..."
glance ./
echo "Glance directory overviews updated successfully."
