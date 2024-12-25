#!/bin/bash
# run `docker run -it -v "$PWD:/code" bats/bats:latest /code/tests/test.bats`
setup() {
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" >/dev/null 2>&1 && pwd )"
    PATH="$DIR/..:$PATH"
}


@test "source script and check env vars" {
    run bash -c "source shell/bayesh.bash && [[ -v BAYESH_PWD ]] && [[ -v BAYESH_CMD ]] && [[ -v BAYESH_LAST_HIST ]]"
    [ "$status" -eq 0 ]
}