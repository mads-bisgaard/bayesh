#!/bin/bash

BAYESH_DIR="${BAYESH_DIR:-~/.bayesh}"
export BAYESH_DIR

# shellcheck source=shell/bayesh.bash
source "$(dirname "${BASH_SOURCE[0]}")/bayesh.bash"


if [[ -z "$PROMPT_COMMAND" ]]; then
    PROMPT_COMMAND="bayesh_update;"
else
    PROMPT_COMMAND="${PROMPT_COMMAND%;}; bayesh_update;"
fi
export PROMPT_COMMAND
export HISTCONTROL=""


__infer_cmd__() {
    local result
    result=$(bayesh_infer_cmd)
    READLINE_LINE="${result}"
    READLINE_POINT=${#result}
}


bind -x '"\C-e":"__infer_cmd__"'
