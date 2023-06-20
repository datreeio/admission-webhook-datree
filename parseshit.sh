#!/bin/bash

# use it like this like this: ` bash parseshit.sh yishay.logs`

# Check if the file path is provided
if [ $# -eq 0 ]; then
  echo "Please provide the path of the file."
  exit 1``
fi

# Read the file path from the command line argument
file_path=$1

# Check if the file exists
if [ ! -f "$file_path" ]; then
  echo "File not found: $file_path"
  exit 1
fi

# Loop over each string in the file
while IFS= read -r input; do
  # Parse the input string using jq
  msg=$(echo "$input" | jq -r '.msg | fromjson')

  # Copy parsed "msg" JSON to clipboard
  if command -v pbcopy &>/dev/null; then
    echo "$msg" | pbcopy
  elif command -v xsel &>/dev/null; then
    echo "$msg" | xsel --clipboard
  else
    echo "Clipboard command not found. Please install 'pbcopy' on macOS or 'xsel' on Linux."
    exit 1
  fi

  # Print the parsed "msg" JSON
  echo "$msg"

  echo "----------------------"
done < "$file_path"
