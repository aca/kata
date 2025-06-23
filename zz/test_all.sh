#!/usr/bin/env bash
echo "=== Argument test ==="
echo "Got $# arguments:"
for i in "$@"; do
    printf "[%s]\n" "$i"
done