package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const (
	goVersion           = "1.21.6"
	golangciLintVersion = "v1.55.2"

	cerbosVersion         = "0.33.0"
	dockerRegistryVersion = "2.8.3"

	batsVersion = "v1.10.0"

	skopeoImageRepo = "quay.io/skopeo/stable"
	skopeoVersion   = "v1.14.1"

	regclientImageRepo = "ghcr.io/regclient/regctl"
	regclientVersion   = "v0.5.6"
)

type Ci struct{}

func (m *Ci) Test(ctx context.Context) (*Container, error) {
	dir := projectDir()
	app := dir.DockerBuild().
		WithMountedFile("/etc/portward/config.yaml", dir.File("./config.ci.yaml")).
		WithMountedFile("/private_key.pem", dir.File("./private_key.pem")).
		WithServiceBinding("cerbos", cerbos()).
		WithExposedPort(8080).
		WithExec([]string{"portward", "--addr", "0.0.0.0:8080", "--debug", "--realm", "localhost:8080", "--config", "/etc/portward/config.yaml"}).
		AsService()

	di, err := dummyImage(ctx)
	if err != nil {
		return nil, err
	}

	return dag.Bats().FromContainer(
		testContainer().
			Container().
			WithMountedFile("/src/var/image.tar.gz", di).
			WithServiceBinding("portward", app).
			WithServiceBinding("registry", dockerRegistry()).
			WithEnvVariable("REGISTRY", "registry:5000"),
	).Run([]string{"bats", "-r", "e2e"}), nil
}

func cerbos() *Service {
	config := dag.Host().Directory(filepath.Join(root(), "etc/cerbos/policies"))

	return dag.Container().From(fmt.Sprintf("ghcr.io/cerbos/cerbos:%s", cerbosVersion)).
		WithExposedPort(3592).
		WithExposedPort(3593).
		WithMountedDirectory("/policies", config).
		AsService()
}

func dockerRegistry() *Service {
	config := dag.Host().Directory(filepath.Join(root(), "etc/docker"))

	return dag.Container().From(fmt.Sprintf("registry:%s", dockerRegistryVersion)).
		WithExposedPort(5000).
		WithMountedDirectory("/etc/docker/registry", config).
		WithEnvVariable("REGISTRY_AUTH_TOKEN_REALM", "http://portward:8080/token").
		AsService()
}

func testContainer() *BatsBaseWithSource {
	// Skopeo needs C libraries, so we install it with apk for now
	// skopeo := client.Container().From(fmt.Sprintf("%s:%s", skopeoImageRepo, skopeoVersion)).File("/usr/bin/skopeo")

	regctl := dag.Container().From(fmt.Sprintf("%s:%s", regclientImageRepo, regclientVersion)).File("/regctl")

	return dag.Bats().FromContainer(
		dag.Bats().
			FromVersion(batsVersion).
			Container().
			WithEntrypoint(nil).
			// WithFile("/usr/bin/skopeo", skopeo).
			WithFile("/usr/bin/regctl", regctl).
			WithExec([]string{"apk", "add", "skopeo"}),
	).WithSource(projectDir())
}

func dummyContainer() *Container {
	return dag.Container(ContainerOpts{Platform: "linux/amd64"}).
		// From("scratch").
		WithNewFile("/hello", ContainerWithNewFileOpts{
			Contents: "hello world",
		})
}

func dummyImage(ctx context.Context) (*File, error) {
	imagePath := filepath.Join(os.TempDir(), "image.tar.gz")
	_, err := dummyContainer().Export(ctx, imagePath)
	if err != nil {
		return nil, err
	}

	return dag.Host().File(imagePath), nil
}

func (m *Ci) Lint() *Container {
	return dag.GolangciLint().
		Run(GolangciLintRunOpts{
			Version:   golangciLintVersion,
			GoVersion: goVersion,
			Source:    projectDir(),
			Verbose:   true,
		})
}
