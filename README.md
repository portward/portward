<p align="center">
  <a href="https://twirphp.github.io">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="resources/logo-dark.png">
      <img alt="Portward logo" src="resources/logo.png" height="300">
    </picture>
  </a>

  <h1 align="center">
    Portward
  </h1>
</p>

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/portward/portward/ci.yaml?style=flat-square)](https://github.com/portward/portward/actions/workflows/ci.yaml)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/mod/github.com/portward/portward)
[![built with nix](https://img.shields.io/badge/builtwith-nix-7d81f7?style=flat-square)](https://builtwithnix.org)

**An all-in-one registry authorization service implementing the [Docker (Distribution) Registry Auth specification](https://github.com/distribution/distribution/tree/main/docs/spec/auth).**

> [!WARNING]
> **Project is under development. Backwards compatibility is not guaranteed.**

## Quickstart

Check out the [quickstart](https://github.com/portward/quickstart/) repository for a demonstration of Portward's capabilities.

## Development

**For an optimal developer experience, it is recommended to install [Nix](https://nixos.org/download.html) and [direnv](https://direnv.net/docs/installation.html).**

### Using Dagger

Run tests:

```shell
dagger call test
```

Run linters:

```shell
dagger call lint
```

### Manual workflow

Launch dependencies:

```shell
docker compose up -d
```

Run the service

```shell
just run
```

Run tests

```shell
just test-all
```

Cleanup:

```shell
docker compose down -v
```

## License

The project is licensed under the [MIT License](LICENSE).
