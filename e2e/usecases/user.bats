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

    _login user password $REGISTRY
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

@test "users can inspect repositories in the root" {
    inspect $REGISTRY/image
}

@test "users cannot inspect repositories they do not have access to" {
    run inspect $REGISTRY/path/to/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"

    run inspect $REGISTRY/producr1/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"
}

@test "users can pull from repositories in the root" {
    pull $REGISTRY/image
}

@test "users cannot pull from repositories they do not have access to" {
    run pull $REGISTRY/path/to/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"

    run pull $REGISTRY/product1/image
    assert_failure
    assert_output --regexp "requested access to the resource is denied|unauthorized"
}

@test "users can push to repositories in their namespace" {
    push $TEST_IMAGE_FILE $REGISTRY/user/image
}

@test "users cannot push to any repositories outside of their namespace" {
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
