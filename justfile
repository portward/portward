default:
    just --list

build:
    mkdir -p build
    go build -o build/portward .

run:
    go run . -addr 0.0.0.0:8080 -debug -realm localhost:8080

test:
    go test -race -v ./...

download-alpine:
    mkdir -p var
    skopeo --insecure-policy copy -a docker://docker.io/library/alpine:latest oci-archive://$PWD/var/alpine.tar.gz

login:
    skopeo --debug login --tls-verify=false -u user -p password 127.0.0.1:5000
    skopeo --debug login --tls-verify=false -u user -p password 127.0.0.1:5001

logout:
    skopeo --debug logout 127.0.0.1:5000
    skopeo --debug logout 127.0.0.1:5001

test-push:
    skopeo --debug --insecure-policy copy --dest-tls-verify=false -a oci-archive://$PWD/var/alpine.tar.gz docker://127.0.0.1:5000/user/alpine
    skopeo --debug --insecure-policy copy --dest-tls-verify=false -a oci-archive://$PWD/var/alpine.tar.gz docker://127.0.0.1:5001/user/alpine

test-push-deny:
    skopeo --debug --insecure-policy copy --dest-tls-verify=false -a oci-archive://$PWD/var/alpine.tar.gz docker://127.0.0.1:5000/alpine
    skopeo --debug --insecure-policy copy --dest-tls-verify=false -a oci-archive://$PWD/var/alpine.tar.gz docker://127.0.0.1:5001/alpine
