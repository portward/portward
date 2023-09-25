# TODO: prepare images in setup (maybe setup suite) for users to pull (currently there is a dependency between tests)
# TODO: cleanup images after

setup_file() {
    bats_load_library bats-support
    bats_load_library bats-assert

    export DOCKER_CONFIG="$(mktemp -d)/auth.json"
    export REGISTRY_AUTH_FILE="$(mktemp -d)/auth.json"

    export REGISTRY="${REGISTRY:-127.0.0.1:5000}"
    export CLIENT="${CLIENT:-skopeo}"

    load ../clients/$CLIENT

    export TEST_IMAGE_FILE="$PWD/var/image.tar.gz"

    _login admin password $REGISTRY
    push $TEST_IMAGE_FILE $REGISTRY/image
    push $TEST_IMAGE_FILE $REGISTRY/path/to/image
    push $TEST_IMAGE_FILE $REGISTRY/product1/image
    _logout $REGISTRY

    _login customer password $REGISTRY
}

setup() {
    bats_load_library bats-support
    bats_load_library bats-assert

    load ../clients/$CLIENT
}

teardown_file() {
    _logout $REGISTRY

    rm -rf $DOCKER_CONFIG
    rm -rf $REGISTRY_AUTH_FILE
}

@test "customers can inspect selected repository" {
    inspect $REGISTRY/product1/image
}

@test "customers cannot inspect repositories they are not entitled to" {
    run inspect $REGISTRY/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"

    run inspect $REGISTRY/path/to/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"
}

@test "customers can pull from selected repository" {
    pull $REGISTRY/product1/image
}

@test "customers cannot pull from repositories they are not entitled to" {
    run pull $REGISTRY/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"

    run pull $REGISTRY/path/to/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"
}

@test "customers cannot push to any repositories" {
    run push $TEST_IMAGE_FILE $REGISTRY/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"

    run push $TEST_IMAGE_FILE $REGISTRY/path/to/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"

    run push $TEST_IMAGE_FILE $REGISTRY/product1/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"
}

# vim:syntax=bash filetype=bash
