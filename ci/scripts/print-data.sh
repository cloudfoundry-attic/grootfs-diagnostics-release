#!/bin/bash
set -e

mkdir -p out
tar xzf diagnostics/*.tar.gz -C out
cd out/*/

echo "d-state processes"
echo "--------------------------------------------"
cat dstate-processes
echo
echo

echo "application.json"
echo "--------------------------------------------"
cat applications.json
echo
echo
