#!/bin/bash

break_script() {
  echo "broken" >&2
  trap - SIGINT
  kill -- -$$
}

terminate_script() {
  echo "killed" >&2
  trap - SIGTERM
  kill -- -$$
}

PARAM="$1"

trap broken_script SIGINT
trap kill_script SIGTERM

echo "starting smart-function-1"

if [ -n "$SF_INPUT_PATH" ] || [ -n "$PARAM" ]; then
  INPUT_FILE="$SF_INPUT_PATH/$PARAM"

  echo "smart-function-1 reading from input file: $INPUT_FILE"
fi

if [ ! -f "$INPUT_FILE" ]; then
  echo "Error: No such file [$INPUT_FILE]" >&2
  exit 1
fi

if [ -n "$SF_OUTPUT_PATH" ]; then
  mkdir -p "$SF_OUTPUT_PATH"
  OUTPUT_FILE="$SF_OUTPUT_PATH/result.txt"
  truncate -s 0 "$OUTPUT_FILE"
  chmod 0666 "$OUTPUT_FILE"

  echo "smart-function-1 output to: $OUTPUT_FILE"
fi

if [ -z "$OUTPUT_FILE" ]; then
  while read -r line; do
    echo "$line" | rev
  done <"$INPUT_FILE"
else
  while read -r line; do
    echo "$line" | rev >>"$OUTPUT_FILE"
  done <"$INPUT_FILE"
fi

exit 0
