#!/bin/bash

FILES=$(find . -type f -name "*.go")
FAIL=0
for file in  $FILES; do
    if [ $(head -n1 "${file}"|grep package|wc -l) -eq 0 ]; then
        echo "${file} starts with package, probably missing copyright"
        FAIL=1
    fi
done

if [ $FAIL -eq 1 ]; then
    echo "missing copyright! fix it!"
    exit 1
fi
exit 0
