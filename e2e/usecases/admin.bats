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

    _login admin password $REGISTRY

    export TEST_IMAGE_FILE="$PWD/var/image.tar.gz"
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

@test "admins can push to any repository" {
    push $TEST_IMAGE_FILE $REGISTRY/image
    push $TEST_IMAGE_FILE $REGISTRY/path/to/image
    push $TEST_IMAGE_FILE $REGISTRY/product1/image
}

@test "admins can inspect any repository" {
    inspect $REGISTRY/image
    inspect $REGISTRY/path/to/image
    inspect $REGISTRY/product1/image
}

@test "admins can pull from any repository" {
    pull $REGISTRY/image
    pull $REGISTRY/path/to/image
    pull $REGISTRY/product1/image
}

# vim:syntax=bash filetype=bash
