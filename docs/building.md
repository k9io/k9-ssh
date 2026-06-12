# Building from Source

k9-ssh is a standard Go module. Building from source requires Go 1.22 or later.

## Quick Build

```bash
git clone https://github.com/k9io/k9-ssh.git
cd k9-ssh
go build -o k9-ssh .
```

The resulting `k9-ssh` binary is statically linked (`CGO_ENABLED=0` is the default for pure-Go builds) and has no runtime dependencies on system libraries.

## Using the Build Script

The included `build.sh` script embeds the current git tag as the version string:

```bash
./build.sh
```

This is equivalent to:

```bash
VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
go build -ldflags "-X main.version=${VERSION}" -o k9-ssh .
```

## Cross-Platform Builds

`scripts/build-all` compiles k9-ssh for every supported OS and architecture combination, producing gzip-compressed binaries and SHA-256 checksums in the `bin/` directory.

```bash
./scripts/build-all
```

Output layout:

```
bin/
├── linux/
│   ├── k9-ssh.amd64.gz
│   ├── k9-ssh.amd64.gz-sha256.txt
│   ├── k9-ssh.arm64.gz
│   ├── k9-ssh.arm64.gz-sha256.txt
│   └── ...
├── freebsd/
│   ├── k9-ssh.amd64.gz
│   └── ...
├── openbsd/
│   └── ...
└── ...
```

### Supported OS / Architecture Matrix

| OS | Architectures |
|----|--------------|
| Linux | amd64, arm64, i386, armv6, armv7, riscv64, mips, mipsle, mips64, mips64le, s390x, ppc64le |
| FreeBSD | amd64, arm64, i386 |
| OpenBSD | amd64, arm64 |
| NetBSD | amd64, arm64 |
| Solaris | amd64 |
| macOS (darwin) | amd64, arm64 |

## Running Tests

```bash
go test ./...
```

Tests cover configuration loading and validation, API response parsing, username validation, and key format validation. All tests run without network access — API calls are mocked via an embedded `httptest` server.

## Dependencies

| Module | Version | Purpose |
|--------|---------|---------|
| `golang.org/x/crypto` | v0.52.0+ | SSH key parsing and validation |
| `gopkg.in/yaml.v3` | v3.0.1+ | YAML configuration parsing |
| `golang.org/x/sys` | v0.45.0+ | Transitive (required by x/crypto) |

Fetch all dependencies with:

```bash
go mod tidy
```

## Verifying a Release Binary

Every release binary ships with a SHA-256 checksum file. Verify before installing:

```bash
# Linux example
sha256sum -c k9-ssh.amd64.gz-sha256.txt

# macOS
shasum -a 256 -c k9-ssh.amd64.gz-sha256.txt

# FreeBSD / OpenBSD
sha256 -c k9-ssh.amd64.gz-sha256.txt
```
