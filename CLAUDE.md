# CLAUDE.md - Container Images Repository

Project-specific instructions for working with this container images repository.

## Repository Context

This repository contains container images optimized for Kubernetes, building upon Alpine or Ubuntu base images with a focus on rootless containers, semantic versioning, and immutable filesystems.

**Key Technologies:**
- Docker / Docker Buildx with Bake
- GitHub Actions for CI/CD
- Go 1.25 for testing (container tests with testcontainers-go)
- Just for task automation
- Renovate for automated dependency updates
- Alpine 3.22 / Python 3.13-alpine3.22 base images

**Go Module:**
- Module path: `github.com/craquehouse/container-images`
- Root-level `go.mod` and `go.sum` required for tests
- Uses testcontainers-go for container testing

## Project Structure

```
.
├── apps/                    # Container application definitions
│   └── {app-name}/
│       ├── Dockerfile       # Container build definition
│       ├── docker-bake.hcl  # Docker Bake configuration
│       ├── container_test.go # Go-based container tests
│       ├── entrypoint.sh    # Container entrypoint script
│       └── scripts/         # Application scripts
├── include/                 # Shared files copied to apps during build
├── testhelpers/            # Go test helper utilities
│   └── testhelpers.go      # Shared test helpers (HTTP, file, command tests)
├── .github/
│   ├── actions/            # Reusable GitHub Actions
│   └── workflows/          # CI/CD workflows
├── .justfile               # Task automation recipes
├── go.mod                  # Go module definition
├── go.sum                  # Go dependency checksums
└── renovate.json5          # Renovate configuration
```

## Development Workflow

### Adding New Container Applications

1. **Directory Structure**: Create `apps/{app-name}/` with required files
   - `Dockerfile` - Alpine 3.22 or Python 3.13-alpine3.22 base
   - `docker-bake.hcl` - Build configuration with APP and VERSION variables
   - `container_test.go` - Go tests using testhelpers
   - `entrypoint.sh` - Container entrypoint (if needed)
   - `scripts/` - Application-specific scripts

2. **Container Requirements**:
   - Must run as `nobody:nogroup` (UID/GID 65534)
   - Support immutable filesystem
   - Use `/data` and/or `/config` for persistent volumes
   - Include catatonit for proper init
   - Set `TERM=xterm-256color`

3. **Testing**: Write Go tests in `container_test.go` using testhelpers package
   - Package name: `main`
   - Import: `"github.com/craquehouse/container-images/testhelpers"`
   - Use `testhelpers.GetTestImage()` to get the test image (respects TEST_IMAGE env var)

### Building & Testing

**Local Build & Test:**
```bash
just local-build {app-name}
```
This recipe:
1. Syncs include/ and apps/{app-name}/ to .cache/
2. Builds the container using docker buildx bake
3. Runs Go tests with TEST_IMAGE environment variable set to the built image

**Remote Build:**
```bash
just remote-build {app-name} [release=true|false]
```

**Prerequisites:**
- Go module must be initialized (`go.mod` and `go.sum` in repository root)
- Run `go mod tidy` after adding/updating dependencies

## Container Image Standards

### Base Image Requirements
- **Alpine**: Must use version 3.22
- **Python**: Must use version 3.13-alpine3.22
- Defined in [renovate.json5](renovate.json5) packageRules

### Container Principles
1. Rootless execution (nobody:nogroup)
2. No s6-overlay usage
3. Semantic versioning in docker-bake.hcl
4. Immutable filesystem support
5. Standardized volume paths: `/data`, `/config`

### Required Packages
- `ca-certificates` - SSL/TLS certificates
- `catatonit` - Proper init system
- Additional packages as needed per application

## File Conventions

### Dockerfile
- Start with `FROM docker.io/library/alpine:3.22` or approved Python base
- Set `ENV TERM=xterm-256color`
- Install packages with `apk add --no-cache`
- Copy entrypoint with `--chmod=755`
- Copy app files with `--chown=65534:65534` (nobody:nogroup)
- Final `USER nobody:nogroup`
- `WORKDIR /app`

### docker-bake.hcl
```hcl
variable "APP" { default = "app-name" }
variable "VERSION" { default = "x.y.z" }
```
- Include annotations for Renovate dependency tracking
- Format: `# renovate: datasource=X depName=Y`
- Define `image`, `image-local`, `image-all` targets
- Multi-platform: linux/amd64, linux/arm64

### Entrypoint Scripts
- Bash shebang: `#!/usr/bin/env bash`
- Set error handling: `set -euo pipefail`
- Use gum for formatted output
- Execute with catatonit: `ENTRYPOINT ["/usr/bin/catatonit", "--", "/entrypoint.sh"]`

### Container Tests (container_test.go)
```go
package main

import (
    "context"
    "testing"
    "github.com/craquehouse/container-images/testhelpers"
)

func Test(t *testing.T) {
    ctx := context.Background()
    image := testhelpers.GetTestImage("ghcr.io/craquehouse/{app}:rolling")

    t.Run("Test Name", func(t *testing.T) {
        // Use testhelpers functions
    })
}
```

**Available testhelpers:**
- `GetTestImage(defaultImage)` - Gets test image from TEST_IMAGE env or uses default
- `TestHTTPEndpoint(t, ctx, image, httpConfig, containerConfig)` - Tests HTTP endpoints
- `TestFileExists(t, ctx, image, filePath, config)` - Verifies file exists
- `TestCommandSucceeds(t, ctx, image, config, entrypoint, args...)` - Tests command execution

**ContainerConfig options:**
- `Env map[string]string` - Environment variables for the container

## Automation & CI/CD

### GitHub Actions Workflows
- **app-builder.yaml**: Builds container images
- **release.yaml**: Publishes releases with semantic versioning
- **pull-request.yaml**: PR validation and testing
- **vulnerability-scan.yaml**: Security scanning
- **test-version.yaml**: Version testing

### Renovate Configuration
- Auto-merges application updates in docker-bake.hcl
- Enforces base image version constraints
- Custom regex manager for docker-bake.hcl annotations
- Semantic commit types: `release(app-name): depName (oldVer → newVer)`

### Just Recipes
- `just local-build {app}` - Build and test locally using .cache directory
- `just remote-build {app} [release]` - Trigger GitHub workflow

## Security & Quality Standards

### Security Requirements
1. Run as non-root user (nobody:nogroup)
2. No unnecessary privileges
3. Minimal package installation
4. Regular vulnerability scanning via GitHub Actions
5. Automated dependency updates via Renovate

### Code Quality
- ShellCheck for shell scripts (`.shellcheckrc` configured)
- Go tests must pass before merge
- EditorConfig for consistent formatting
- GitHub CodeQL analysis enabled

## Deprecation Policy

Containers may be deprecated when:
1. Upstream application no longer maintained
2. Official upstream container meets project goals
3. Better alternative exists
4. Maintenance burden too high

**Deprecated containers remain published for 6 months before pruning.**

## Tool Configuration

### mise (.mise.toml)
Managed development tools:
- just 1.43.1
- gh (GitHub CLI) 2.83.1
- jq 1.7.1
- yq 4.49.2
- go 1.25

### EditorConfig
- Indent: 2 spaces
- Trim trailing whitespace
- Insert final newline
- UTF-8 encoding

## Common Tasks

### Creating New App
1. Create `apps/{name}/` directory
2. Add Dockerfile with Alpine 3.22 base
3. Create docker-bake.hcl with APP and VERSION
4. Write container_test.go (package main, import testhelpers)
5. Add entrypoint.sh if needed
6. Update `go.mod` if needed: `go mod tidy`
7. Test locally: `just local-build {name}`
8. Create PR with changes

### Updating App Version
1. Update VERSION in docker-bake.hcl
2. Update dependencies via Renovate annotations
3. If Go dependencies changed: `go mod tidy`
4. Run tests: `just local-build {name}`
5. Renovate auto-creates PR for dependency updates

### Adding Dependencies
1. Add Renovate annotation in docker-bake.hcl:
   ```hcl
   # renovate: datasource=github-releases depName=owner/repo
   variable "DEP_VERSION" { default = "1.2.3" }
   ```
2. Use variable in Dockerfile or scripts

## Best Practices

1. **Always read existing apps** before creating new ones to follow patterns
2. **Test locally first** using `just local-build` before pushing
3. **Keep Dockerfiles minimal** - only install necessary packages
4. **Document environment variables** in comments
5. **Use semantic versioning** strictly in docker-bake.hcl
6. **Leverage testhelpers** for consistent Go testing
7. **Follow shellcheck** recommendations for scripts
8. **Maintain immutability** - avoid writing to filesystem outside volumes
9. **Run `go mod tidy`** after modifying test dependencies
10. **Use testcontainers-go** patterns from testhelpers for container tests

## References

- Inspired by: [bjw-s-labs/container-images](https://github.com/bjw-s-labs/container-images)
- GitHub Packages: [craquehouse packages](https://github.com/orgs/craquehouse/packages?tab=packages&repo_name=container-images)
- Security: See [.github/SECURITY.md](.github/SECURITY.md)

---

*Container Images Repository | Kubernetes-optimized containers | Rootless, semantic-versioned, immutable*
