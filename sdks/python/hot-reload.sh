#!/bin/sh

if [ -z "$1" ]; then
  script="simple"
else
  # Reject input that contains shell metacharacters
  case "$1" in
    *\;*|*\&*|*\|*|*\<*|*\>*|*\(*|*\)*|*\{*|*\}*|*\`*|*\\*|*\$*)
      echo "Error: Invalid script name provided. The script name should not contain shell metacharacters."
      exit 1
      ;;
    *)
      script="$1"
      ;;
  esac
fi

watchmedo auto-restart --recursive --patterns="*.py" -- poetry run "$script"