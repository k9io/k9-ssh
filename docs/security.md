# Security

k9-ssh is designed around a defense-in-depth model. This page describes the security controls built into the program and the recommended hardening steps for your deployment.

## Built-In Security Controls

### Dedicated Unprivileged User

k9-ssh enforces that it runs as the user specified in `run_as` (typically `key9`). If the process user at runtime does not match, k9-ssh logs the mismatch to syslog and exits with a non-zero status. This prevents privilege escalation if the binary is somehow called by an unexpected user or process.

The `key9` account should be configured with no home directory, no login shell, no password, and no sudo access.

### Username Validation

Every username passed via `--user` is validated against the pattern `^[a-z_][a-z0-9_\-]{0,31}$` before the API is contacted. This prevents:

- Path traversal (e.g., `../../etc/passwd` would be rejected immediately)
- Shell injection via username
- API endpoint manipulation

If validation fails, k9-ssh exits silently (no API call, no key output) and writes a warning to syslog.

### Public Key Validation

Every key returned by the Key9 API is parsed through Go's `golang.org/x/crypto/ssh` library before being printed to stdout. Keys that fail to parse or are of an unrecognized type are silently dropped and logged. This ensures sshd never receives malformed data that could exploit a parsing bug in older OpenSSH versions.

### API Communication

By default, all communication with the Key9 cloud API uses HTTPS. When [k9-proxy](k9-proxy.md) is deployed as a local cache on the same host, k9-ssh can use a plain `http://` URL to reach it because traffic is confined to the loopback interface and never crosses the network. TLS is only required when k9-proxy is serving multiple hosts over a network.

### Fail-Safe Error Handling

On any error — network failure, non-200 status, JSON parse error, API error response — k9-ssh logs the problem to syslog and exits cleanly with no output. From sshd's perspective, no keys means authentication falls through to the next method. This prevents transient errors from granting inadvertent access and avoids leaking diagnostic information to the connecting user.

### Audit Logging

All operations are logged to syslog using the `AUTH` facility and `INFO` priority under the tag `k9-ssh`. Logged events include:

| Event | Log message |
|-------|------------|
| Invalid username | Warning with the rejected value |
| API connection error | Error with URL and Go error string |
| Non-200 HTTP status | Status code received |
| JSON parse error | Error from Go's JSON decoder |
| API-returned error | Error string from Key9 |
| Key validation failure | Rejected key fingerprint |
| Process user mismatch | Expected vs. actual user |

On most systems these logs appear in `/var/log/auth.log` (Debian/Ubuntu), `/var/log/secure` (RHEL/Fedora), or are readable via `journalctl -t k9-ssh`.

### Static Binaries

k9-ssh is compiled with `CGO_ENABLED=0`, producing a fully static binary with no dependency on system C libraries (glibc, musl, etc.). This eliminates an entire class of library-based vulnerabilities and makes the binary identical across installations of the same OS/arch.

## Recommended Hardening

### File Permissions

```bash
# Binary: root-owned, world-executable (required by sshd)
sudo chown root:root /opt/k9/bin/k9-ssh
sudo chmod 755 /opt/k9/bin/k9-ssh

# Config: readable only by root and the key9 group
sudo chown root:key9 /opt/k9/etc/k9.yaml
sudo chmod 640 /opt/k9/etc/k9.yaml

# Config directory: not world-readable
sudo chmod 750 /opt/k9/etc
```

### Restrict key9 Account

```bash
# Verify no login shell
grep key9 /etc/passwd
# Should show: /bin/false or /sbin/nologin

# Verify no sudo access
sudo -l -U key9
# Should show: not allowed to run sudo
```

### Disable Password Authentication

Once k9-ssh is working, disable password-based SSH login:

```
# /etc/ssh/sshd_config
PasswordAuthentication no
KbdInteractiveAuthentication no
```

### Disable Local authorized_keys (Optional)

To make Key9 the sole source of authorized keys:

```
# /etc/ssh/sshd_config
AuthorizedKeysFile none
```

This prevents local `~/.ssh/authorized_keys` files from bypassing Key9 access controls.

### Network Egress

k9-ssh only needs outbound HTTPS access to the Key9 API host (`ssh-api.k9.io`, port 443). If you run a host-based firewall, restrict outbound connections from the `key9` user accordingly:

**iptables (Linux)**

```bash
# Allow key9 user to reach the Key9 API
iptables -A OUTPUT -m owner --uid-owner key9 -d ssh-api.k9.io -p tcp --dport 443 -j ACCEPT
# Block everything else from key9
iptables -A OUTPUT -m owner --uid-owner key9 -j DROP
```

**pf (OpenBSD / FreeBSD)**

```
pass out on egress proto tcp from any to <k9-api-hosts> port 443 user key9
block out on egress from any to any user key9
```

### API Key Rotation

Rotate your Key9 API key periodically and after any suspected compromise. After updating the key in the Key9 dashboard:

1. Update `/opt/k9/etc/k9.yaml` with the new key.
2. Reload sshd if it caches the key (typically it does not, since k9-ssh is re-executed per login).
3. Verify with `sudo -u key9 /opt/k9/bin/k9-ssh --user=<test-user>`.

## Threat Model

| Threat | Mitigation |
|--------|-----------|
| Attacker supplies crafted username to manipulate API URL | Username regex rejects anything that is not a valid Unix username |
| Malicious API response contains invalid key data | All keys are validated before output |
| k9-ssh called by unexpected user | `run_as` check causes immediate exit |
| API credentials extracted from config | File permissions (640, root:key9); config not readable by world |
| Network interception of API traffic | HTTPS enforced for cloud API; plain HTTP acceptable only for loopback-bound k9-proxy (traffic never leaves the host) |
| k9-ssh binary tampered with | SHA-256 checksums provided for all release binaries; install from trusted source only |
| Key9 service unavailable | Fail-safe: no keys returned means authentication falls through; combine with a local break-glass account if needed |
