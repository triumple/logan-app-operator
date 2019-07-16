#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

echo "--- cpuinfo ---"
cat /proc/cpuinfo
echo "--- meminfo ---"
cat /proc/meminfo
echo "--- system ulimit ---"
ulimit -a