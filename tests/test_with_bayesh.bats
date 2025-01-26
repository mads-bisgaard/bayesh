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

@test "check only record new commands" {
    source shell/bayesh.bash
    command="random command ${RANDOM}"
    history -s "${command}"
    db=$(bayesh print-settings | jq -r .db)
    
    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '0'
    
    # allow time for insertion into db. why doesn't 'wait' work?
    _bayesh_update && sleep 1
    
    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '1'
    run bash -c "sqlite3 ${db} 'select event_counter from events'"
    [ "$status" -eq 0 ]
    assert_output '1'

    _bayesh_update && sleep 1
    
    run bash -c "sqlite3 ${db} 'select count(*) from events'"
    [ "$status" -eq 0 ]
    assert_output '1'
    run bash -c "sqlite3 ${db} 'select event_counter from events'"
    [ "$status" -eq 0 ]
    assert_output '1'    
}
