#!/bin/bash
# tests must be run from the root directory of the repo

setup_file() {

    # install and expose bayesh on PATH
    venv="$(mktemp -d)/.venv"
    # shellcheck source=/dev/null
    python -m venv "${venv}" && source "${venv}/bin/activate"
    python -m pip install .
    ln -s "${venv}/bin/bayesh" "/usr/local/bin/bayesh"
    bayesh --help
}

setup() {
    bats_load_library bats-support
    bats_load_library bats-assert
    BAYESH_DIR=$(mktemp -d)
    export BAYESH_DIR
}

teardown() {
    rm -rf "${BAYESH_DIR}"
}

@test "source script and check env vars" {
    run bash -c \
    '
    source shell/bayesh.bash
    [[ -v BAYESH_PWD ]] && [[ -v BAYESH_CMD ]] && [[ -v BAYESH_LAST_HIST ]]
    '
    [ "$status" -eq 0 ]
}

@test "test _bayesh_post_process_command 3 tokens" {
    run bash -c \
    '
    source shell/bayesh.bash
    _bayesh_post_process_command "This is a test <ABC> string with <DEF> multiple <XYZ> entries."
    '
    assert_output '15
This is a test  string with  multiple  entries.'
    [ "$status" -eq 0 ]
}

@test "test _bayesh_post_process_command 0 tokens" {
    run bash -c \
    '
    source shell/bayesh.bash
    _bayesh_post_process_command "This is a test"
    '
    assert_output '14
This is a test'
    [ "$status" -eq 0 ]
}