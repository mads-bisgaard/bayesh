#!/usr/bin/env bash
# End-to-end tests of the zsh integration, driving interactive shells in a
# private tmux server.
bats_require_minimum_version 1.5.0

setup_file() {
    command -v bayesh &> /dev/null
    command -v tmux &> /dev/null || skip "tmux is not installed"
    command -v fzf &> /dev/null || skip "fzf is not installed"
}

setup() {
    load /batslib/bats-support/load
    load /batslib/bats-assert/load

    TEST_DIR=$(mktemp -d)
    export TEST_DIR
    mkdir -p "${TEST_DIR}/zdot/completions" "${TEST_DIR}/tmp" "${TEST_DIR}/bayesh" "${TEST_DIR}/work"
    export SOCKET="bayesh-test-$$-${BATS_TEST_NUMBER}"

    cat > "${TEST_DIR}/zdot/completions/_fpathtool" <<'EOF'
#compdef fpathtool
compadd alpha beta
EOF
    cat > "${TEST_DIR}/zdot/rc-common" <<'EOF'
HISTSIZE=1000
PROMPT='%% '
source <(bayesh --zsh)
EOF
    cat > "${TEST_DIR}/zdot/.zshrc" <<'EOF'
fpath=($ZDOTDIR/completions $fpath)
autoload -Uz compinit
compinit -d $ZDOTDIR/.zcompdump
_fntool() { compadd fn_one fn_two }
compdef _fntool fntool
_slowtool() { sleep 3; compadd slow_one }
compdef _slowtool slowtool
source $ZDOTDIR/rc-common
EOF
}

teardown() {
    tmux -L "${SOCKET}" kill-server 2> /dev/null || true
    rm -rf "${TEST_DIR}"
}

t() {
    tmux -L "${SOCKET}" "$@"
}

start_tmux() {
    env ZDOTDIR="${TEST_DIR}/zdot" BAYESH_DIR="${TEST_DIR}/bayesh" \
        TMPDIR="${TEST_DIR}/tmp" \
        tmux -L "${SOCKET}" -f /dev/null new-session -d -s bayesh -n one \
        -x 120 -y 30 -c "${TEST_DIR}/work" "zsh -i"
    wait_for_pane bayesh:one.0 '^%'
}

# wait_until <command...>: polls for up to 8 seconds
wait_until() {
    local _
    for _ in $(seq 1 160); do
        "$@" && return 0
        sleep 0.05
    done
    echo "timed out waiting for: $*" >&2
    return 1
}

pane_matches() {
    t capture-pane -p -t "$1" 2> /dev/null | grep -qE -- "$2"
}

wait_for_pane() {
    wait_until pane_matches "$1" "$2" || {
        echo "--- pane $1:" >&2
        t capture-pane -p -t "$1" >&2
        return 1
    }
}

# prompt_is <pane> <regex>: the line being edited matches the regex
prompt_is() {
    t capture-pane -p -t "$1" 2> /dev/null | grep -v '^$' | tail -n 1 | grep -qE -- "$2"
}

pane_count() {
    [[ $(t list-panes -t "$1" 2> /dev/null | wc -l) -eq "$2" ]]
}

window_gone() {
    ! t list-windows -t bayesh -F '#{window_name}' 2> /dev/null | grep -qx "$1"
}

# run_cmd <window> <command>: run a command in the shell and wait for the prompt
run_cmd() {
    t send-keys -t "bayesh:$1.0" -l "$2"
    t send-keys -t "bayesh:$1.0" Enter
    sleep 0.2
}

open_bayesh() {
    t send-keys -t "bayesh:$1.0" C-e
    wait_until pane_count "bayesh:$1" 2
    wait_for_pane "bayesh:$1.1" "$2"
}

# events: the recorded events as "previous -> current", oldest first
events() {
    sqlite3 "${TEST_DIR}/bayesh/bayesh.db" \
        "select previous_cmd || ' -> ' || current_cmd from events order by last_modified" 2> /dev/null
}

events_equal() {
    [[ "$(events)" == "$1" ]]
}

# events_are <expected>: waits for the recorded events to be exactly <expected>
events_are() {
    wait_until events_equal "$1" || {
        echo "--- events:" >&2
        events >&2
        return 1
    }
}

@test "commands are recorded with the previous command" {
    start_tmux
    run_cmd one 'echo one'
    run_cmd one 'echo two'

    events_are ' -> echo one
echo one -> echo two'
}

@test "an empty line is not recorded" {
    start_tmux
    run_cmd one 'echo one'
    t send-keys -t bayesh:one.0 Enter
    sleep 0.2
    run_cmd one 'echo two'

    events_are ' -> echo one
echo one -> echo two'
}

@test "a multi-line command is recorded as typed" {
    start_tmux
    t send-keys -t bayesh:one.0 -l "echo 'a"
    t send-keys -t bayesh:one.0 Enter
    run_cmd one "b'"

    events_are " -> echo 'a
b'"
}

@test "commands from other shells' shared history are not recorded" {
    echo "HISTFILE=\$ZDOTDIR/.zsh_history; SAVEHIST=1000; setopt share_history" >> "${TEST_DIR}/zdot/rc-common"
    start_tmux
    t new-window -t bayesh -n two -c "${TEST_DIR}/work" "zsh -i"
    wait_for_pane bayesh:two.0 '^%'
    run_cmd one 'echo first'
    run_cmd two 'echo a'
    run_cmd two 'echo b'
    t send-keys -t bayesh:one.0 Enter
    sleep 0.2
    run_cmd one 'echo after'

    events_are ' -> echo first
 -> echo a
echo a -> echo b
echo first -> echo after'
}

@test "repeated commands are recorded with hist_ignore_dups" {
    echo "setopt hist_ignore_all_dups" >> "${TEST_DIR}/zdot/rc-common"
    start_tmux
    run_cmd one 'echo one'
    run_cmd one 'echo one'

    events_are ' -> echo one
echo one -> echo one'
}

@test "commands kept out of history with hist_ignore_space are not recorded" {
    echo "setopt hist_ignore_space" >> "${TEST_DIR}/zdot/rc-common"
    start_tmux
    run_cmd one 'echo one'
    run_cmd one ' echo secret'
    run_cmd one 'echo two'

    events_are ' -> echo one
echo one -> echo two'
}

@test "the recording state is not exported to child processes" {
    start_tmux
    run_cmd one 'echo one'
    run_cmd one "env > ${TEST_DIR}/env"
    wait_until test -s "${TEST_DIR}/env"

    run grep -E '^_?BAYESH_' "${TEST_DIR}/env"
    assert_output "BAYESH_DIR=${TEST_DIR}/bayesh"
}

@test "the bayesh pane opens to the right of the shell" {
    start_tmux
    open_bayesh one '│'

    run t display-message -p -t bayesh:one.1 '#{pane_at_right} #{pane_at_top} #{pane_at_bottom}'
    assert_output '1 1 1'
    run t display-message -p -t bayesh:one.0 '#{pane_at_left} #{pane_at_right}'
    assert_output '1 0'
    # a side panel, narrower than the shell
    (( $(t display-message -p -t bayesh:one.1 '#{pane_width}') * 100 <= 35 * $(t display-message -p -t bayesh:one '#{window_width}') ))
}

@test "selecting a suggestion puts it in the prompt" {
    start_tmux
    run_cmd one 'echo one'
    run_cmd one 'echo one'
    open_bayesh one 'echo one'

    t send-keys -t bayesh:one.0 C-Right

    wait_until prompt_is bayesh:one.0 '^% echo one$'
}

@test "Ctrl-Left offers completions for the current buffer" {
    start_tmux
    open_bayesh one '│'
    t send-keys -t bayesh:one.0 -l 'fntool f'

    t send-keys -t bayesh:one.0 C-Left

    wait_for_pane bayesh:one.1 'fntool fn_one'
}

@test "Ctrl-Left completes at the cursor" {
    start_tmux
    open_bayesh one '│'
    t send-keys -t bayesh:one.0 -l 'fntool f --flag'
    t send-keys -t bayesh:one.0 Left Left Left Left Left Left Left

    t send-keys -t bayesh:one.0 C-Left

    wait_for_pane bayesh:one.1 'fntool fn_one --flag'
}

@test "Ctrl-Left leaves the prompt unchanged" {
    start_tmux
    open_bayesh one '│'
    t send-keys -t bayesh:one.0 -l 'fntool f'

    t send-keys -t bayesh:one.0 C-Left
    wait_for_pane bayesh:one.1 'fntool fn_one'

    prompt_is bayesh:one.0 '^% fntool f$'
}

@test "selecting a suggestion offers completions from the user's shell" {
    start_tmux
    run_cmd one 'fpathtool a'
    run_cmd one 'fpathtool a'
    open_bayesh one 'fpathtool a'

    t send-keys -t bayesh:one.0 C-Right

    wait_for_pane bayesh:one.1 'fpathtool alpha'
}

@test "completion functions defined in the shell are used" {
    start_tmux
    run_cmd one 'fntool f'
    run_cmd one 'fntool f'
    open_bayesh one 'fntool f'

    t send-keys -t bayesh:one.0 C-Right

    wait_for_pane bayesh:one.1 'fntool fn_one'
}

@test "completion works without compinit in the user's configuration" {
    cp "${TEST_DIR}/zdot/rc-common" "${TEST_DIR}/zdot/.zshrc"
    # some distributions run compinit in /etc/zsh/zshrc
    echo 'skip_global_compinit=1' > "${TEST_DIR}/zdot/.zshenv"
    touch "${TEST_DIR}/work/somefile"
    start_tmux
    run_cmd one 'ls some'
    run_cmd one 'ls some'
    open_bayesh one 'ls some'

    t send-keys -t bayesh:one.0 C-Right

    wait_for_pane bayesh:one.1 'ls somefile'
    run t send-keys -t bayesh:one.0 C-u
    # shellcheck disable=SC2016
    t send-keys -t bayesh:one.0 -l 'echo "comps=${+_comps}"'
    t send-keys -t bayesh:one.0 Enter
    wait_for_pane bayesh:one.0 'comps=0'
    [[ ! -e "${TEST_DIR}/zdot/.zcompdump" ]]
}

# Completers, matchers and file tags are tried in turn until one of them adds a match
assert_single_completion() {
    run_cmd one "$1"
    run_cmd one "$1"
    open_bayesh one "$1"

    t send-keys -t bayesh:one.0 C-Right
    wait_for_pane bayesh:one.1 "$2"
    sleep 0.3

    run grep -c -- "$2" <(t capture-pane -p -t bayesh:one.1)
    assert_output '1'
}

@test "a completion is listed once with several matchers" {
    echo "zstyle ':completion:*' matcher-list '' 'm:{a-z}={A-Z}'" >> "${TEST_DIR}/zdot/.zshrc"
    touch "${TEST_DIR}/work/somefile"
    start_tmux
    assert_single_completion 'ls some' 'ls somefile'
}

@test "a completion is listed once with several completers" {
    echo "zstyle ':completion:*' completer _complete _approximate" >> "${TEST_DIR}/zdot/.zshrc"
    touch "${TEST_DIR}/work/somefile"
    start_tmux
    assert_single_completion 'ls some' 'ls somefile'
}

@test "a completion is listed once when files are matched by several tags" {
    touch "${TEST_DIR}/work/archive.tar"
    start_tmux
    assert_single_completion 'tar xf arch' 'tar xf archive.tar'
}

@test "completion happens at the cursor, not at the end of the line" {
    start_tmux
    run_cmd one 'fpathtool "x y" --flag'
    run_cmd one 'fpathtool "x y" --flag'
    open_bayesh one 'fpathtool'

    t send-keys -t bayesh:one.0 C-Right

    wait_for_pane bayesh:one.1 'fpathtool "alpha.*--flag'
}

@test "the prompt stays responsive during a slow completion" {
    start_tmux
    run_cmd one 'slowtool s'
    run_cmd one 'slowtool s'
    open_bayesh one 'slowtool s'

    t send-keys -t bayesh:one.0 C-Right
    sleep 0.3
    t send-keys -t bayesh:one.0 -l 'typed'
    sleep 0.5

    pane_matches bayesh:one.0 'slowtool styped'
}

@test "replacing an unfinished completion leaves no fifo behind" {
    start_tmux
    run_cmd one 'slowtool s'
    run_cmd one 'slowtool s'
    open_bayesh one 'slowtool s'

    t send-keys -t bayesh:one.0 C-Right
    sleep 0.3
    t send-keys -t bayesh:one.0 C-c
    sleep 0.5
    run_cmd one 'echo done'
    sleep 0.5

    run find "${TEST_DIR}/tmp" -type p
    assert_output ''
}

@test "a completion without matches does not ring the bell" {
    start_tmux
    run_cmd one 'ls /nonexistent/'
    run_cmd one 'ls /nonexistent/'
    open_bayesh one 'ls /nonexistent/'
    t set-option -g bell-action any
    t set-window-option -g monitor-bell on

    t send-keys -t bayesh:one.0 C-Right
    sleep 0.5

    run t display-message -p -t bayesh:one '#{window_bell_flag}'
    assert_output '0'
}

@test "multibyte input reaches fzf intact" {
    start_tmux
    run_cmd one 'echo æ'
    run_cmd one 'echo æøå'
    run_cmd one 'true' # a third echo would fold the two into echo <ARG>
    open_bayesh one 'echo'

    t send-keys -t bayesh:one.0 -l 'echo æøå'

    wait_until bash -c "! tmux -L ${SOCKET} capture-pane -p -t bayesh:one.1 | grep -qE 'echo æ +│'"
    pane_matches bayesh:one.1 'echo æøå'
}

@test "a closing parenthesis in the buffer reaches fzf intact" {
    start_tmux
    run_cmd one 'echo one'
    run_cmd one 'echo two'
    open_bayesh one 'echo'

    t send-keys -t bayesh:one.0 -l 'echo two)'

    wait_until bash -c "! tmux -L ${SOCKET} capture-pane -p -t bayesh:one.1 | grep -q 'echo'"
}

@test "each shell has its own bayesh session" {
    start_tmux
    t new-window -t bayesh -n two -c "${TEST_DIR}/work" "zsh -i"
    wait_for_pane bayesh:two.0 '^%'
    run_cmd one 'echo from-one'
    run_cmd one 'echo from-one'
    run_cmd two 'echo from-two'
    run_cmd two 'echo from-two'

    open_bayesh one 'echo'
    open_bayesh two 'echo'

    t send-keys -t bayesh:one.0 C-e
    wait_until pane_count bayesh:one 1
    pane_count bayesh:two 2
}

@test "exiting a shell closes its bayesh pane only" {
    start_tmux
    t new-window -t bayesh -n two -c "${TEST_DIR}/work" "zsh -i"
    wait_for_pane bayesh:two.0 '^%'
    open_bayesh one '│'
    open_bayesh two '│'

    run_cmd one 'exit'

    wait_until window_gone one
    pane_count bayesh:two 2
}

@test "killing a shell closes its bayesh pane" {
    start_tmux
    t new-window -t bayesh -n two -c "${TEST_DIR}/work" "zsh -i"
    wait_for_pane bayesh:two.0 '^%'
    run_cmd one 'echo one'
    run_cmd one 'echo one'
    open_bayesh one 'echo one'

    t kill-pane -t bayesh:one.0

    wait_until window_gone one
}

@test "a shell started from a bayesh shell does not touch the parent's session" {
    start_tmux
    open_bayesh one '│'

    run_cmd one 'zsh -i'
    sleep 0.5
    run_cmd one 'exit'
    sleep 0.5

    pane_count bayesh:one 2
}

@test "a shell survives its bayesh pane dying and can reopen it" {
    start_tmux
    open_bayesh one '│'

    t kill-pane -t bayesh:one.1
    wait_until pane_count bayesh:one 1
    t send-keys -t bayesh:one.0 C-e

    wait_until pane_count bayesh:one 2
}

@test "the integration does not depend on curl" {
    mkdir "${TEST_DIR}/shim"
    printf '#!/bin/sh\necho called >> "%s/curl-calls"\n' "${TEST_DIR}" > "${TEST_DIR}/shim/curl"
    chmod +x "${TEST_DIR}/shim/curl"
    PATH="${TEST_DIR}/shim:${PATH}" start_tmux
    run_cmd one 'echo one'
    run_cmd one 'echo one'
    open_bayesh one 'echo one'

    t send-keys -t bayesh:one.0 -l 'echo'
    sleep 0.3

    [[ ! -e "${TEST_DIR}/curl-calls" ]]
}

# run_loop <window>: runs a for loop typed over several lines
run_loop() {
    t send-keys -t "bayesh:$1.0" -l 'for f in x'
    t send-keys -t "bayesh:$1.0" Enter
    t send-keys -t "bayesh:$1.0" -l 'do true'
    t send-keys -t "bayesh:$1.0" Enter
    run_cmd "$1" 'done'
}

# last_event: the most recently recorded command
last_event() {
    sqlite3 "${TEST_DIR}/bayesh/bayesh.db" \
        "select current_cmd from events order by last_modified desc limit 1" 2> /dev/null
}

last_event_is() {
    [[ "$(last_event)" == "$1" ]]
}

@test "a multi-line suggestion is a single item" {
    start_tmux
    run_loop one
    run_cmd one 'echo one'
    run_loop one
    events_are ' -> for f in x
do true
done
for f in x
do true
done -> echo one
echo one -> for f in x
do true
done'
    # echo one follows the loop, so it is ranked first
    open_bayesh one 'echo one'
    pane_matches bayesh:one.1 'do true'

    t send-keys -t bayesh:one.0 C-Down
    t send-keys -t bayesh:one.0 C-Right
    sleep 0.3
    t send-keys -t bayesh:one.0 Enter

    events_are ' -> for f in x
do true
done
for f in x
do true
done -> echo one
echo one -> for f in x
do true
done
for f in x
do true
done -> for f in x
do true
done'
}

@test "Ctrl-Left offers completions for a multi-line buffer" {
    start_tmux
    open_bayesh one '│'
    t send-keys -t bayesh:one.0 -l 'echo x'
    t send-keys -t bayesh:one.0 M-Enter
    t send-keys -t bayesh:one.0 -l 'fntool f'

    t send-keys -t bayesh:one.0 C-Left
    wait_for_pane bayesh:one.1 'fntool fn_one'
    t send-keys -t bayesh:one.0 C-Right
    sleep 0.3
    t send-keys -t bayesh:one.0 C-e Enter

    wait_until last_event_is 'echo x
fntool fn_one'
}

@test "completing in the middle of a multi-line buffer keeps the other lines" {
    start_tmux
    open_bayesh one '│'
    t send-keys -t bayesh:one.0 -l 'fntool f'
    t send-keys -t bayesh:one.0 M-Enter
    t send-keys -t bayesh:one.0 -l 'echo x'
    # the cursor goes to the end of the first line
    t send-keys -t bayesh:one.0 C-a Left

    t send-keys -t bayesh:one.0 C-Left
    wait_for_pane bayesh:one.1 'fntool fn_one'
    t send-keys -t bayesh:one.0 C-Right
    sleep 0.3
    t send-keys -t bayesh:one.0 C-e Enter

    wait_until last_event_is 'fntool fn_one
echo x'
}

# is_current <pane> <text>: the line showing <text> has fzf's pointer
is_current() {
    t capture-pane -p -t "$1" | grep -F -- "$2" | grep -q '▌'
}

# stripe_of <pane> <text>: the background escape of the line showing <text>,
# or nothing when it has the default background
stripe_of() {
    t capture-pane -p -e -t "$1" | grep -F -- "$2" | grep -oE '\[48;5;[0-9]+m' | head -n 1
}

@test "every line of a multi-line suggestion shares one stripe" {
    start_tmux
    run_loop one
    run_cmd one 'echo one'
    run_loop one
    run_cmd one 'echo two'
    open_bayesh one 'echo two'
    # Moves the highlight off the loop, the first item, to echo two, the third
    t send-keys -t bayesh:one.0 C-Down
    t send-keys -t bayesh:one.0 C-Down
    wait_until is_current bayesh:one.1 'echo two'

    # Items alternate between no stripe and a stripe, so the lines of the loop
    # have none, where striping every line would give 'do true' one
    [[ -z "$(stripe_of bayesh:one.1 'for f in x')" ]]
    [[ -z "$(stripe_of bayesh:one.1 'do true')" ]]
    [[ -z "$(stripe_of bayesh:one.1 'done')" ]]
    [[ -n "$(stripe_of bayesh:one.1 'echo one')" ]]

    # The stripe stands out from the terminal background, and the current item
    # stands out from the stripe
    [[ "$(stripe_of bayesh:one.1 'echo one')" == '[48;5;237m' ]]
    [[ "$(stripe_of bayesh:one.1 'echo two')" == '[48;5;239m' ]]
}
