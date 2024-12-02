#!/bin/bash

BAYESH_DIR="${BAYESH_DIR:-~/.bayesh}"
export BAYESH_DIR

if [[ -z "${BAYESH_SRC_DIR}" ]]; then
    echo "Error: BAYESH_SRC_DIR is not defined." >&2
    return 1
elif [[ ! -d "${BAYESH_SRC_DIR}" ]]; then
    echo "Error: BAYESH_SRC_DIR is not a directory." >&2
    return 1
fi

function bayesh() {
    "${BAYESH_SRC_DIR}"/.venv/bin/python3 -m bayesh "$@"
}


# shellcheck source=shell/bayesh.bash
source "$(dirname "${BASH_SOURCE[0]}")/bayesh.bash"


if [[ -z "$PROMPT_COMMAND" ]]; then
    PROMPT_COMMAND="bayesh_update;"
else
    PROMPT_COMMAND="${PROMPT_COMMAND%;}; bayesh_update;"
fi
export PROMPT_COMMAND
export HISTCONTROL=""

#####################################

# __infer_cmd__() {
#     local result
#     result=$(bayesh_infer_cmd)
#     READLINE_LINE="${result}"
#     READLINE_POINT=${#result}
# }


# bind -x '"\C-e":"__infer_cmd__"'
