#!/usr/bin/env zsh

autoload -Uz add-zsh-hook
# add hook for updating bayesh db
add-zsh-hook precmd _bayesh_update

if [[ ! -n "$TMUX" ]]; then

    function bayesh_infer_cmd() {
        local cur
        local result
        local point
        
        cur="${CURSOR}"
        result=$(_bayesh_infer_cmd)
        line=$(echo "${result}" | tail -n 1);point=$(echo "${result}" | head -n 1)
        LBUFFER="${LBUFFER}${line}${RBUFFER}"
        zle reset-prompt
        CURSOR=$(( cur + point ))
    }
    zle -N infer_cmd bayesh_infer_cmd
    bindkey '^E' infer_cmd # Ctrl-e    
    return

fi

function _bayesh_is_active() {
    [[ -n "$BAYESH_SERVER_CONFIG" ]] || return 1
    client_pid=$(echo "$BAYESH_SERVER_CONFIG" | jq -r .client_pid)
    [[ "$client_pid" == "$$" ]] || return 1
}


#functions for communicating with server

function bayesh_start_or_kill_server() {

    if _bayesh_is_active; then
        if _fzf_tmux_server_get -c "$BAYESH_SERVER_CONFIG" &> /dev/null; then
            _fzf_tmux_server_kill -c "$BAYESH_SERVER_CONFIG"
            unset BAYESH_SERVER_CONFIG
            return 
        fi
    fi
    config=$(_fzf_tmux_server_start)
    BAYESH_SERVER_CONFIG=$(echo "$config" | jq -Mc ". + { \"client_pid\": $$ }")
    export BAYESH_SERVER_CONFIG
    zle-line-init
}
zle -N start_or_kill_server bayesh_start_or_kill_server
bindkey '^E' start_or_kill_server


function zle-line-init() {
    if _bayesh_is_active; then
        fifo=$(mktemp -u)
        mkfifo "$fifo"
        (
            bayesh infer-cmd "$(pwd)" "${BAYESH_CMD}" > "$fifo" &
            echo "search()" | _fzf_tmux_server_post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null &
            echo "reload(cat $fifo; rm $fifo)" | _fzf_tmux_server_post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null &
        )
    fi
}
zle -N zle-line-init


function zle-line-pre-redraw() {
    if _bayesh_is_active; then
        ( echo "search("$BUFFER")" | _fzf_tmux_server_post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null & )
    fi
}
zle -N zle-line-pre-redraw


function bayesh_select() {
    local token_regex
    local cmd
    local p
    token_regex="<STRING>|<PATH>"

    if _bayesh_is_active; then
        cmd=$(_fzf_tmux_server_get -c "$BAYESH_SERVER_CONFIG" 2> /dev/null | jq -r .current.text)
        p="${#cmd}"
        if echo "${cmd}" | grep -boq -E "${token_regex}"; then
            p=$(echo "${cmd}" | grep -bo -E "${token_regex}" | cut -d: -f1 | head -n1)
        fi
        BUFFER=$(echo "${cmd}" | sed -E "s/(${token_regex})//g")
        zle -R
        CURSOR="$p"
    fi    
}
zle -N select bayesh_select
bindkey '^[[1;5C' select # Ctrl-rightarrow

function bayesh_up() {
    if _bayesh_is_active; then
        ( echo "up" | _fzf_tmux_server_post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null & )
    fi    
}
zle -N up bayesh_up
bindkey '^[[1;5A' up # Ctrl-uparrow

function bayesh_down() {
    if _bayesh_is_active; then
        ( echo "down" | _fzf_tmux_server_post -c "$BAYESH_SERVER_CONFIG" 2> /dev/null & )
    fi    
}
zle -N down bayesh_down
bindkey '^[[1;5B' down # Ctrl-downarrow

trap '_fzf_tmux_server_kill -c "$BAYESH_SERVER_CONFIG" &> /dev/null' EXIT
