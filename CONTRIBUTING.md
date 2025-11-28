# Contributing to uniget

Thank you for considering contributing to uniget! We welcome any contributions, whether it's bug fixes, new features, or improvements to the existing codebase.

## Contribution Prerequisites

Make sure that you have [Docker](https://www.docker.com/) and [buildx](https://github.com/docker/buildx) installed. You can use a locally installed [Go](https://go.dev/) but the containerized build environment enables you to use the same tooling as in CI.

## Sending a Pull Request

1. Create an issue in the repository outlining the fix or feature
1. Fork the repository to your own GitLab account and clone it locally
1. Complete and test the change
1. Add tests for new code
1. Create a concise commit message and reference the issue(s) and pull request(s) adressed
1. Ensure that CI passes. If it fails, fix the failures
1. Every pull request requires a review

The following steps describe the build, linting and testing processes:

### Building uniget CLI

To build locally, run this command:

```shell
docker buildx bake binary
```

### Running tests

To run the tests, execute the following commands:

```shell
docker buildx bake lint
docker buildx bake vet
docker buildx bake gosec
docker buildx bake test
```

**Make sure all tests pass** without any failures or errors.

## What to contribute

### Did one of the installed tools misbehave?

* Please refer to the [tools repository](https://gitlab.com/uniget-org/backlog/-/issues)

### Do you have a feature suggestion?

* **Ensure the feature was not already suggested** by searching on GitLab under [Issues](https://gitlab.com/uniget-org/cli/issues?q=is%3Aopen+is%3Aissue+label%3Aenhancement)

* If you're unable to find an open issue describing the feature, [open a new suggestion](https://gitlab.com/uniget-org/backlog/-/issues/new)

### Did you find a bug?

* **Ensure the bug was not already reported** by searching on GitLab under [Issues](https://gitlab.com/uniget-org/backlog/-/issues)

* If you're unable to find an open issue addressing the problem, [open a new one](https://gitlab.com/uniget-org/backlog/-/issues/new)

### Did you write a patch that fixes a bug?

* Open a new GitLab merge request with the patch

* Ensure the MR description clearly describes the problem and solution. Include the relevant issue number if applicable

### Do you intend to add a new feature or change an existing one?

* Open a new GitLab merge request with the code
* Ensure the PR description clearly describes the feature and the implementation. Include the relevant issue number if applicable

### Did you fix whitespace, format code, or make a purely cosmetic patch?

* Please treat this like a feature request and follow the steps above

## Code of Conduct

uniget adheres to and enforces the [Contributor Covenant](http://contributor-covenant.org/version/1/4/) Code of Conduct.
Please take a moment to read the [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) document.
