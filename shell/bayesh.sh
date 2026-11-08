#!/bin/sh

# Lets the user pick a suggestion for the command following $1 with fzf, and
# prints the cursor position followed by the chosen command without tokens.
_bayesh_infer_cmd() {
    token_regex="<ARG>"
    chosen_cmd=$( 
    inferred_cmds=$(bayesh infer-cmd "$(pwd)" "$1")
    
    if [ "${BAYESH_AVOID_IF_EMPTY+set}" ] && [ -z "$(echo "${inferred_cmds}" | awk '{$1=$1};1')" ]; then
        exit 1
    fi
    echo "${inferred_cmds}" | fzf \
        --scheme=history \
        --no-sort \
        --exact \
        --bind="zero:print-query" \
        --bind="ctrl-q:print-query" \
        --ansi \
        --preview="echo {} | sed -E 's/(${token_regex})//g'" \
        --border=none \
        --preview-window=border-rounded,up:1:wrap \
        --header="Press Ctrl+q to select query" \
        --header-first \
        --info=inline-right \
        --layout=reverse \
        --margin=0 \
        --padding=0 \
        --height=30% \
        --no-mouse
    ) || return

    position="${#chosen_cmd}"
    if echo "${chosen_cmd}" | grep -boq -E "${token_regex}"; then
        position=$(echo "${chosen_cmd}" | grep -bo -E "${token_regex}" | cut -d: -f1 | head -n1)
    fi
    echo "${position}"
    echo "${chosen_cmd}" | sed -E "s/(${token_regex})//g"
}
