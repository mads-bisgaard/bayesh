#!/usr/bin/env bash
# tests must be run from the root directory of the repo
bats_require_minimum_version 1.5.0

setup_file() {
    command -v bayesh &> /dev/null
}

setup() {
    load /batslib/bats-support/load
    load /batslib/bats-assert/load
    BAYESH_DIR=$(mktemp -d)
    export BAYESH_DIR
}

teardown() {
    rm -rf "${BAYESH_DIR}"
}

@test "test bayesh installed" {
    run bayesh --help
    [ "$status" -eq 0 ]
}

@test "test inference function (no tokens)" {
    db=$(bayesh settings | jq -r .BAYESH_DATABASE)

    cwd=$(mktemp -d)
    previous_cmd="previous command ${RANDOM}"
    current_cmd="current command ${RANDOM}"
    bayesh record-event "${cwd}" "${previous_cmd}" "${current_cmd}"

    run sqlite3 "${db}" 'select count(*) from events'
    assert_output '1'

    # the picker used outside tmux
    export cwd previous_cmd
    export FZF_DEFAULT_OPTS="--filter=\"${current_cmd}\""
    run env -u TMUX zsh -i -c 'source <(bayesh --zsh); cd "${cwd}"; _bayesh_infer_cmd "${previous_cmd}"' </dev/null
    [ "$status" -eq 0 ]
    assert_output "${#current_cmd}
${current_cmd}"
}


@test "test bayesh server active function" {
    run zsh -i -c \
    '
    set -e; \
    export TMUX="tmuxsessionid"; \
    source <(bayesh --zsh); \
    _bayesh_is_active;
    '
    [ "$status" -eq 1 ]


    run zsh -i -c \
    '
    set -e; \
    export TMUX="tmuxsessionid"; \
    source <(bayesh --zsh); \
    export BAYESH_SERVER_CONFIG="{\"url\":\"http://localhost:8000\", \"tmux_pane_id\": \"%1\", \"client_pid\": \"not a pid\"}"; \
    _bayesh_is_active;
    '
    [ "$status" -eq 1 ]


    run zsh -i -c \
    '
    set -e; \
    export TMUX="tmuxsessionid"; \
    source <(bayesh --zsh); \
    export BAYESH_SERVER_CONFIG="{\"url\":\"http://localhost:8000\", \"tmux_pane_id\": \"%1\", \"client_pid\":$$}"; \
    _bayesh_is_active;
    '
    [ "$status" -eq 0 ]    
}
@test "record-event stores its arguments as given" {
    db=$(bayesh settings | jq -r .BAYESH_DATABASE)

    # an empty previous command, as for the first command of a shell
    run bayesh record-event /work "" "echo one"
    [ "$status" -eq 0 ]
    # commands which look like flags, are empty or have surrounding whitespace
    run bayesh record-event /work "--help" " -n x "
    [ "$status" -eq 0 ]

    run sqlite3 "${db}" "select quote(cwd), quote(previous_cmd), quote(current_cmd) from events order by last_modified"
    assert_output "'/work'|''|'echo one'
'/work'|'--help'|' -n x '"
}

@test "infer-cmd takes an empty or flag-like previous command" {
    bayesh record-event /work "" "echo one"
    bayesh record-event /work "-x" "echo two"

    run bayesh infer-cmd /work ""
    [ "$status" -eq 0 ]
    assert_output --partial "echo one"
    run bayesh infer-cmd /work "-x"
    [ "$status" -eq 0 ]
    assert_line --index 0 "echo two"
}
