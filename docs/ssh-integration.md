# SSH Integration

k9-ssh integrates with OpenSSH through the `AuthorizedKeysCommand` directive introduced in OpenSSH 6.2. This page explains how the integration works and how to configure it.

## How AuthorizedKeysCommand Works

When a user attempts to log in via SSH, sshd normally checks `~/.ssh/authorized_keys`. If `AuthorizedKeysCommand` is configured, sshd additionally calls the specified program and treats its stdout as a supplemental list of authorized public keys. If the user's private key matches any key from either source, authentication succeeds.

k9-ssh implements this interface. sshd calls k9-ssh with the connecting username, k9-ssh fetches that user's registered keys from Key9, validates each key's format, and prints valid keys one per line to stdout.

## sshd_config Directives

### Basic Setup

```
AuthorizedKeysCommand /opt/k9/bin/k9-ssh --user=%u
AuthorizedKeysCommandUser key9
```

`%u` is an sshd token that is replaced with the username of the connecting user at runtime.

`AuthorizedKeysCommandUser` specifies the OS user that sshd uses to execute the command. This must match the `run_as` value in `k9.yaml`. Using a dedicated unprivileged user ensures k9-ssh cannot be exploited to gain elevated privileges.

### With Client Address (OpenSSH 9.4+)

```
AuthorizedKeysCommand /opt/k9/bin/k9-ssh --user=%u --remote=%C
AuthorizedKeysCommandUser key9
```

The `%C` token expands to a string identifying the client connection (typically `address:port:lport:rdomain`). When provided, k9-ssh forwards this information to the Key9 API, enabling IP-based policy enforcement on the Key9 side.

To check your OpenSSH version:

```bash
ssh -V
```

### Full sshd_config Example

```
# Key9 SSH key retrieval
AuthorizedKeysCommand /opt/k9/bin/k9-ssh --user=%u --remote=%C
AuthorizedKeysCommandUser key9

# Standard settings (adjust to your policy)
PasswordAuthentication no
ChallengeResponseAuthentication no
PubkeyAuthentication yes
```

> Setting `PasswordAuthentication no` after verifying k9-ssh works ensures that only Key9-managed keys can be used to log in.

## sshd Token Reference

The tokens below can be used in the `AuthorizedKeysCommand` value. k9-ssh uses `%u` and optionally `%C`.

| Token | Expands to |
|-------|-----------|
| `%u` | Username of the connecting user |
| `%C` | Client connection string (`address:port:lport:rdomain`) — OpenSSH 9.4+ |
| `%h` | Home directory of the connecting user |
| `%f` | Fingerprint of the key being offered (not useful for k9-ssh) |

## API Request Flow

For each login attempt, k9-ssh makes one HTTPS POST request to the Key9 API:

```
POST {query_ssh_keys}{username}/{machine_group}
Authorization: API_KEY: {company_uuid}:{api_key}
Content-Type: application/json

{"remote": "{client-address}"}   ← only if --remote was provided
```

The response is a newline-delimited stream of JSON objects, one per key:

```json
{"public_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... user@host"}
{"public_key": "sk-ssh-ed25519@openssh.com AAAAGnNrL..."}
```

Error responses use:

```json
{"error": "user not found in machine group"}
```

k9-ssh validates each returned key with Go's `golang.org/x/crypto/ssh` library before printing it. Malformed or unrecognized keys are silently dropped and logged to syslog.

## Username Validation

Before making any API call, k9-ssh validates the username against the pattern:

```
^[a-z_][a-z0-9_\-]{0,31}$
```

This accepts standard Unix usernames (lowercase letters, digits, underscores, hyphens, up to 32 characters). Any username that does not match is rejected with a syslog warning and k9-ssh exits without contacting the API. This prevents path traversal and injection attacks via the username field.

## Coexistence with authorized_keys

k9-ssh and local `authorized_keys` files are not mutually exclusive — sshd evaluates both. If you want Key9 to be the sole source of truth for SSH keys, set the following in `sshd_config`:

```
AuthorizedKeysFile none
```

This disables local `authorized_keys` lookup entirely and forces all authentication through `AuthorizedKeysCommand`.

## Testing Without Restarting sshd

You can test k9-ssh output directly without touching sshd configuration:

```bash
# Run as the key9 user with a known Key9 username
sudo -u key9 /opt/k9/bin/k9-ssh --user=alice

# Test with a custom config path
sudo -u key9 /opt/k9/bin/k9-ssh --user=alice --config=/etc/k9.yaml
```

A successful response looks like:

```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkMfWGH0fCQhkL... alice@laptop
```

No output (but exit code 0) means the user exists but has no keys registered for this machine group. Check Key9 dashboard for the user's group assignments.
