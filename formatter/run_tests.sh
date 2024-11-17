#!/usr/bin/bash

BaseDir=$(dirname "$0")
(
    cd "$BaseDir"
    coverageFile="./coverage.out"
    coverageHtml="./coverage.html"

    [ -f "$coverageFile" ] && rm "$coverageFile"
    [ -f "$coverageHtml" ] && rm "$coverageHtml"

    find ../testdata/cleaned/actual -type f -exec rm {} \;
    find ../testdata/formatted/actual -type f -exec rm {} \;
    find ../testdata/output/actual -type f -exec rm {} \;
    find ../testdata/tagged/actual -type f -exec rm {} \;

    echo ""
    echo "### test:"
    go test -coverprofile="$coverageFile"

    echo ""
    echo "### coverage:"
    go tool cover -func="$coverageFile"
    go tool cover -html="$coverageFile" -o="$coverageHtml"
)
