#!/usr/bin/bash

BaseDir=$(dirname "$0")
(
    cd "$BaseDir"
    coverageFile="./coverage.out"
    coverageHtml="./coverage.html"

    [ -f "$coverageFile" ] && rm "$coverageFile"
    [ -f "$coverageHtml" ] && rm "$coverageHtml"
    find "./testdata/result" -type f -exec rm {} \;

    echo ""
    echo "### gocyclo:"
    gocyclo *.go

    echo ""
    echo "### test:"
    go test -coverprofile="$coverageFile"

    echo ""
    echo "### coverage:"
    go tool cover -func="$coverageFile"
    go tool cover -html="$coverageFile" -o="$coverageHtml"
)
