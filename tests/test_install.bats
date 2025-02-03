#!/bin/bash
# tests must be run from the root directory of the repo
bats_require_minimum_version 1.5.0

_repo=$(pwd)
_repo_copy=$(mktemp -d)

setup() {
    bats_load_library bats-support
    bats_load_library bats-assert
    cp -r "${_repo}" "${_repo_copy}" || exit 1
    cd "${_repo_copy}" || exit 1
    cd "$(ls)" || exit 1
}

teardown() {
    cd "${_repo}" || exit 1
    rm -rf "${_repo_copy}"
    rm -f /usr/local/bin/bayesh
    run -127 bayesh --help
    [ "$status" -eq 127 ]    
}


@test "test install bayesh" {
    run ./install.bash
    [ "$status" -eq 0 ]
    run bayesh --help
    [ "$status" -eq 0 ]
}
