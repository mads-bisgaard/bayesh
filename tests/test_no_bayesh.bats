#!/bin/bash
# tests must be run from the root directory of the repo


setup() {
    bats_load_library bats-support
    bats_load_library bats-assert
}

@test "test source script" {
    run bash -c \
    '
    source shell/bayesh.bash
    [[ -v BAYESH_PWD ]] && [[ -v BAYESH_CMD ]] && [[ -v BAYESH_LAST_HIST ]]
    '
    [ "$status" -eq 0 ]
}

@test "test _bayesh_post_process_command with 3 tokens" {
    run bash -c \
    '
    source shell/bayesh.bash
    _bayesh_post_process_command "This is a test <ABC> string with <DEF> multiple <XYZ> entries."
    '
    assert_output '15
This is a test  string with  multiple  entries.'
    [ "$status" -eq 0 ]
}

@test "test _bayesh_post_process_command with 0 tokens" {
    run bash -c \
    '
    source shell/bayesh.bash
    _bayesh_post_process_command "This is a test"
    '
    assert_output '14
This is a test'
    [ "$status" -eq 0 ]
}