#!/bin/sh
set -eu

export REPORT_OUTPUT="out/e2e_report.html"
rm -f $REPORT_OUTPUT
export E2E_REPORT=1

FORCE_COLOR=1 DEBUG=1 go run ./e2etests/report/main.go "$@";

if [ -z "${NO_OPEN:-}" ]; then
  if [ -s "$REPORT_OUTPUT" ]; then
    if [ -x "$(command -v open)" ]; then
      open $REPORT_OUTPUT
    elif [ -x "$(command -v xdg-open)" ]; then
      xdg-open $REPORT_OUTPUT
    else
      echo "Please open $REPORT_OUTPUT"
    fi
  else
    echo "The report is empty"
  fi
fi
