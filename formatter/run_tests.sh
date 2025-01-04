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
    find ../testdata/untagged/actual -type f -exec rm {} \;

go test


    # echo ""
    # echo "### test:"
    # go test -coverprofile="$coverageFile"
    #
    # echo ""
    # echo "### coverage:"
    # go tool cover -func="$coverageFile"
    # go tool cover -html="$coverageFile" -o="$coverageHtml"

    differs=$(diff -rq ../testdata/output/{actual,expected} | grep differ | wc -l)
    missing=$(diff -rq ../testdata/output/{actual,expected} | grep 'Only in ../testdata/output/actual' | wc -l)
    fcount=$(find ../testdata/output/actual -type f | wc -l)

    ts=$(date +"%F %R")
    pct=$(psql -t -c "select round ( ( 1.0 - ${differs}/(${fcount}-${missing})::numeric ) * 100, 4)")

    echo "# ${ts} ${fcount} files processed"
    echo "# ${ts} ${missing} files not compared (no expected file to compare)"
    echo "# ${ts} ${differs} files differ"
    echo "# ${ts} Pass rate ${pct}%"

meld ../testdata/output/*
)
