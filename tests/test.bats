#!/bin/bash
# From the repo root, run
# docker run -it -v "$PWD:/code" bats/bats:latest tests
setup() {

    bats_load_library bats-support
    bats_load_library bats-assert
}


@test "source script and check env vars" {
    run bash -c \
    '
    source shell/bayesh.bash
    [[ -v BAYESH_PWD ]] && [[ -v BAYESH_CMD ]] && [[ -v BAYESH_LAST_HIST ]]
    '
    [ "$status" -eq 0 ]
}

@test "test bayesh_post_process_command 3 tokens" {
    run bash -c \
    '
    source shell/bayesh.bash
    my_array=()
    bayesh_post_process_command my_array "This is a test <ABC> string with <DEF> multiple <XYZ> entries."
    echo "${my_array[@]}"
    '
    assert_output '15 This is a test  string with  multiple  entries.'
    [ "$status" -eq 0 ]
}

@test "test bayesh_post_process_command 0 tokens" {
    run bash -c \
    '
    source shell/bayesh.bash
    my_array=()
    bayesh_post_process_command my_array "This is a test"
    echo "${my_array[@]}"
    '
    assert_output '14 This is a test'
    [ "$status" -eq 0 ]
}