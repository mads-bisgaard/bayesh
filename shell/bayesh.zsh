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

# BAYESH_SERVER_CONFIG is inherited by child processes, so client_pid is what ties
# a server to the one shell which started it.
function _bayesh_is_active() {
    [[ "$BAYESH_SERVER_CONFIG" == *'"client_pid":'"$$"[,}]* ]]
}

# Lets the server read from a new fifo, replacing whatever it is reading now.
# The fifo is available in $_BAYESH_FIFO, which the caller must start writing to.
# fzf only keeps the latest of its pending searches and reloads, so the query
# is sent along with the reload, where a search sent after it can't drop it.
function _bayesh_begin_reload() {
    _bayesh_end_reload
    _BAYESH_FIFO=$(mktemp -u)
    mkfifo "$_BAYESH_FIFO"
    _BAYESH_QUERY="$LBUFFER"
    _fzf_tmux_server_post "$BAYESH_SERVER_CONFIG" "reload(cat $_BAYESH_FIFO; rm $_BAYESH_FIFO)+search:$_BAYESH_QUERY"
}

# The server removes the fifo once it has been read, so if it is still here the
# writer is unfinished, and its pid has not been reused.
function _bayesh_end_reload() {
    [[ -p "$_BAYESH_FIFO" ]] || return 0
    kill "$_BAYESH_WRITER_PID" 2> /dev/null
    rm -f "$_BAYESH_FIFO"
}

#functions for communicating with server

function bayesh_start_or_kill_server() {

    if _bayesh_is_active; then
        if _fzf_tmux_server_get "$BAYESH_SERVER_CONFIG" &> /dev/null; then
            _bayesh_end_reload
            _fzf_tmux_server_kill "$BAYESH_SERVER_CONFIG"
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
        _fzf_tmux_server_post "$BAYESH_SERVER_CONFIG" "first"
        _bayesh_begin_reload
        bayesh infer-cmd "$PWD" "${BAYESH_CMD}" > "$_BAYESH_FIFO" &!
        _BAYESH_WRITER_PID=$!
    fi
}
zle -N zle-line-init


# Only a changed query is sent, since every search would replace a pending reload
function zle-line-pre-redraw() {
    if _bayesh_is_active && [[ "$LBUFFER" != "$_BAYESH_QUERY" ]]; then
        _BAYESH_QUERY="$LBUFFER"
        _fzf_tmux_server_post "$BAYESH_SERVER_CONFIG" "search:$_BAYESH_QUERY"
    fi
    return 0
}
zle -N zle-line-pre-redraw


function bayesh_select() {
    local cmd
    local token='<(ARG|STRING|PATH)>'

    if _bayesh_is_active; then
        cmd=$(_fzf_tmux_server_get "$BAYESH_SERVER_CONFIG" | jq -r '.current.text // empty')
        [[ -n "$cmd" ]] || return 0
        BUFFER="${cmd//${~token}/}"
        CURSOR=$(( ${cmd[(i)${~token}*]} - 1 ))
        zle -R
        bayesh_complete
    fi
    return 0
}
zle -N select bayesh_select
bindkey '^[[1;5C' select # Ctrl-rightarrow

# Fills the fzf pane with the shell's completions of the buffer at the cursor
function bayesh_complete() {
    if _bayesh_is_active; then
        _bayesh_begin_reload
        zle _bayesh_capture_completions
    fi
    return 0
}
zle -N complete bayesh_complete
bindkey '^[[1;5D' complete # Ctrl-leftarrow

function bayesh_up() {
    if _bayesh_is_active; then
        _fzf_tmux_server_post "$BAYESH_SERVER_CONFIG" up
    fi
    return 0
}
zle -N up bayesh_up
bindkey '^[[1;5A' up # Ctrl-uparrow

function bayesh_down() {
    if _bayesh_is_active; then
        _fzf_tmux_server_post "$BAYESH_SERVER_CONFIG" down
    fi
    return 0
}
zle -N down bayesh_down
bindkey '^[[1;5B' down # Ctrl-downarrow

function _bayesh_exit() {
    if _bayesh_is_active; then
        _bayesh_end_reload
        _fzf_tmux_server_kill "$BAYESH_SERVER_CONFIG" &> /dev/null || true
    fi
}
trap _bayesh_exit EXIT
