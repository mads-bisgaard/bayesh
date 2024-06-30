#!/bin/bash

#set -e
#set -u


function bayesh_record_event() {
    bayesh record-event "${PWD}" "${BAYESH_CMD}" "$0"
}

export BAYESH_DIR="${HOME}/.bayesh"
export BAYESH_CMD=""
export PROMPT_COMMAND=
#'$(bayesh record-event "${PWD}" "${BAYESH_CMD}" "$(fc -ln -1)") && export BAYESH_CMD=$(fc -ln -1)'


if [[ -n "$PROMPT_COMMAND" ]]; then
    PROMPT_COMMAND='$PROMPT_COMMAND; echo $(fc -ln -1)'
else
    PROMPT_COMMAND='echo $(fc -ln -1)'
fi
