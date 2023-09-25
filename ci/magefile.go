//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"dagger.io/dagger"
	_ "github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

const (
	batsImageRepo = "bats/bats"
	batsVersion   = "v1.10.0"

	skopeoImageRepo = "quay.io/skopeo/stable"
	skopeoVersion   = "v1.13.3"

	regclientImageRepo = "ghcr.io/regclient/regctl"
	regclientVersion   = "v0.5.1"
)

func Build(ctx context.Context) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Host().
		Directory(".", dagger.HostDirectoryOpts{
			Exclude: []string{
				"/.devenv/",
				"/.direnv/",
				"/.github/",
				"/bin/",
				"/build/",
				"/ci/",
				"/Dockerfile",
				"/var/",
			},
		}).
		DockerBuild().
		Sync(ctx)
	if err != nil {
		return err
	}

	return nil
}

func Test(ctx context.Context) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	var cerbos *dagger.Container
	var dockerRegistry *dagger.Container

	var test *dagger.Container

	// Prepare
	{

		client := client.Pipeline("Prepare")

		{
			client := client.Pipeline("Cerbos")

			cerbos = cerbosContainer(client)
		}

		{
			client := client.Pipeline("Docker registry")

			dockerRegistry = dockerRegistryContainer(client)
		}
	}

	// Build
	{
		client := client.Pipeline("Build")

		var app *dagger.Container
		{
			client := client.Pipeline("App")
			host := client.Host()

			app = appContainer(client).
				WithMountedFile("/etc/portward/config.yaml", host.File("./config.ci.yaml")).
				WithMountedFile("/private_key.pem", host.File("./private_key.pem")).
				WithServiceBinding("cerbos", cerbos).
				WithExposedPort(8080, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp}).
				WithExec([]string{"portward", "--addr", "0.0.0.0:8080", "--debug", "--realm", "localhost:8080", "--config", "/etc/portward/config.yaml"})
		}

		var dummyImage *dagger.File
		{
			client := client.Pipeline("Test image")
			var err error

			dummyImage, err = buildDummyImage(ctx, client)
			if err != nil {
				return err
			}
		}

		{
			client := client.Pipeline("Test container")

			test = testContainer(client).
				WithMountedFile("/usr/local/src/portward/var/image.tar.gz", dummyImage).
				WithServiceBinding("portward", app).
				WithServiceBinding("registry", dockerRegistry).
				WithEnvVariable("REGISTRY", "registry:5000")
		}
	}

	testContainerID, err := test.ID(ctx)
	if err != nil {
		return err
	}

	_, err = client.Pipeline("Test").
		Container(dagger.ContainerOpts{
			ID: testContainerID,
		}).
		WithFocus().
		WithExec([]string{"bats", "-r", "e2e"}).
		Sync(ctx)
	if err != nil {
		return err
	}

	return err
}

func appContainer(client *dagger.Client) *dagger.Container {
	return client.Host().
		Directory(".", dagger.HostDirectoryOpts{
			Exclude: []string{
				".devenv/",
				".direnv/",
				".github/",
				"bin/",
				"build/",
				"ci/",
				"var/",
			},
		}).
		DockerBuild()
}

func testContainer(client *dagger.Client) *dagger.Container {
	// Skopeo needs C libraries, so we install it with apk for now
	// skopeo := client.Container().From(fmt.Sprintf("%s:%s", skopeoImageRepo, skopeoVersion)).File("/usr/bin/skopeo")

	regctl := client.Container().From(fmt.Sprintf("%s:%s", regclientImageRepo, regclientVersion)).File("/regctl")

	dir := client.Host().
		Directory(".", dagger.HostDirectoryOpts{
			Exclude: []string{
				".devenv/",
				".direnv/",
				".github/",
				"bin/",
				"build/",
				"ci/",
				"Dockerfile",
				"var/",
			},
		})

	return client.Container().
		From(fmt.Sprintf("%s:%s", batsImageRepo, batsVersion)).
		WithEntrypoint(nil).
		// WithFile("/usr/bin/skopeo", skopeo).
		WithFile("/usr/bin/regctl", regctl).
		WithExec([]string{"apk", "add", "skopeo"}).
		WithMountedDirectory("/usr/local/src/portward", dir).
		WithWorkdir("/usr/local/src/portward")
}

func dockerRegistryContainer(client *dagger.Client) *dagger.Container {
	config := client.Host().Directory("./etc/docker")

	return client.Container().From("registry:2.8.2").
		WithExposedPort(5000, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp}).
		WithMountedDirectory("/etc/docker/registry", config).
		WithEnvVariable("REGISTRY_AUTH_TOKEN_REALM", "http://portward:8080/token").
		WithExec(nil)
}

func cerbosContainer(client *dagger.Client) *dagger.Container {
	config := client.Host().Directory("./etc/cerbos/policies")

	return client.Container().From("ghcr.io/cerbos/cerbos:0.30.0").
		WithExposedPort(3592, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp}).
		WithExposedPort(3593, dagger.ContainerWithExposedPortOpts{Protocol: dagger.Tcp}).
		WithMountedDirectory("/policies", config).
		WithExec(nil)
}

func dummyImage(client *dagger.Client) *dagger.Container {
	return client.Container(dagger.ContainerOpts{
		Platform: "linux/amd64",
	}).
		// From("scratch").
		WithNewFile("/hello", dagger.ContainerWithNewFileOpts{
			Contents: "hello world",
		})
}

func buildDummyImage(ctx context.Context, client *dagger.Client) (*dagger.File, error) {
	imagePath := filepath.Join(os.TempDir(), "image.tar.gz")
	_, err := dummyImage(client).Export(ctx, imagePath)
	if err != nil {
		return nil, err
	}

	return client.Host().File(imagePath), nil
}
