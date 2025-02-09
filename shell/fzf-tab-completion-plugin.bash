#!/usr/bin/env bash

function bayesh_autocomplete() {
    local input
    input=$(echo "$READLINE_LINE" | awk '{$1=$1};1')
    if [[ "$input" == "" ]]; then
        bayesh_infer_cmd
    else
        fzf_bash_completion
    fi
}
