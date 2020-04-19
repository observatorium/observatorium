#!/usr/bin/env bash

# This script uses jsonnetfmt to format jsonnet files.

find . -type f -not -path './jsonnet/vendor/*' \( -name '*.libsonnet' -o -name '*.jsonnet' \) | xargs -L1 jsonnetfmt -n 2 --max-blank-lines 2 --string-style s --comment-style s -i

