#!/usr/bin/env bash

name="Merge pull request #44 from triumple/operator_timezone_fix"
re="Merge pull request #(^[0-9]+) .*"
if [[ $name =~ $re ]]; then
 echo ${BASH_REMATCH[1]};
else
 echo ""
fi