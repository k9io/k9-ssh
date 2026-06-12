# k9-ssh

**k9-ssh** is a lightweight SSH public key retrieval agent that bridges OpenSSH with the [Key9](https://k9.io) Identity and Access Management (IAM) platform. It enables centralized, passwordless SSH authentication across your fleet without distributing or managing individual `authorized_keys` files.

## How It Works

OpenSSH's `AuthorizedKeysCommand` directive allows sshd to call an external program at login time to retrieve a user's authorized public keys. k9-ssh implements this interface: when a user attempts to log in, sshd calls k9-ssh with the username, k9-ssh queries the Key9 API, validates each returned key, and prints valid keys to stdout for sshd to evaluate.

```
User SSH Login
     │
     ▼
sshd (AuthorizedKeysCommand)
     │
     ▼
k9-ssh --user=<username> [--remote=<client-info>]
     │
     ├── Load /opt/k9/etc/k9.yaml
     ├── POST https://ssh-api.k9.io/api/v1/ssh/query/<user>/<group>
     └── Validate & print public keys
     │
     ▼
sshd completes authentication
```

## Key Features

- **Centralized key management** — public keys live in Key9, not on individual servers
- **Machine group scoping** — restrict which Key9 users can log in to which groups of machines
- **Fail-safe design** — errors log silently without exposing key material
- **Minimal footprint** — ~670 lines of Go, statically compiled, no libc dependency
- **Full audit trail** — every operation logged to syslog `AUTH` facility
- **Platform portable** — Linux, FreeBSD, OpenBSD, NetBSD, Solaris, and more

## Supported Platforms

k9-ssh ships as statically compiled binaries for all major Unix-like operating systems and architectures supported by Go, including:

- Linux (amd64, arm64, i386, armv6, armv7, riscv64, and more)
- FreeBSD (amd64, arm64, i386)
- OpenBSD (amd64, arm64)
- NetBSD (amd64, arm64)
- Solaris / illumos (amd64)
- macOS / Darwin (amd64, arm64)

## License

k9-ssh is released under the GNU General Public License v2.
