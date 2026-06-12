# Troubleshooting

## Quick Diagnostics

Before diving in, run k9-ssh manually as the `key9` user. This bypasses sshd entirely and shows exactly what k9-ssh returns:

```bash
sudo -u key9 /opt/k9/bin/k9-ssh --user=<username>
```

Then check syslog for any messages from k9-ssh:

```bash
# Debian / Ubuntu
grep k9-ssh /var/log/auth.log

# RHEL / Fedora / AlmaLinux
grep k9-ssh /var/log/secure

# systemd (most Linux)
journalctl -t k9-ssh

# FreeBSD / OpenBSD
grep k9-ssh /var/log/authlog
```

---

## Common Problems

### No keys returned, no syslog errors

**Symptom:** k9-ssh exits with code 0 and no output. sshd falls back to other auth methods (password prompt, or `Permission denied`).

**Causes and checks:**

1. **User not in machine group** — The Key9 user does not have access to the machine group configured in `k9.yaml`. Log in to the Key9 dashboard and verify the user's group assignments match `machine_group` in your config.

2. **User has no SSH keys in Key9** — The user exists in Key9 but has not uploaded an SSH public key. Have them add a key via the Key9 dashboard.

3. **Wrong machine group name** — Typo in `k9.yaml`. Group names are case-sensitive. Check `machine_group` against the exact name shown in the Key9 dashboard.

---

### `user mismatch` in syslog

**Symptom:**
```
k9-ssh: user mismatch: expected key9, got root
```

**Cause:** k9-ssh was called by a user other than the one specified in `run_as`. This almost always means `AuthorizedKeysCommandUser` in `sshd_config` does not match `run_as` in `k9.yaml`.

**Fix:** Ensure both values are identical (typically `key9`):

```
# sshd_config
AuthorizedKeysCommandUser key9
```

```yaml
# k9.yaml
system:
  run_as: "key9"
```

---

### `invalid username` in syslog

**Symptom:**
```
k9-ssh: invalid username: Username123
```

**Cause:** The username passed via `--user` does not match the allowed pattern `^[a-z_][a-z0-9_\-]{0,31}$`. This means:
- The username contains uppercase letters
- The username contains characters other than `a-z`, `0-9`, `_`, or `-`
- The username is longer than 32 characters
- The username starts with a digit

**Fix:** Ensure the Key9 username matches a valid Unix username. Rename the account in Key9 if necessary.

---

### API connection error in syslog

**Symptom:**
```
k9-ssh: error contacting API: Post "https://ssh-api.k9.io/...": dial tcp: i/o timeout
```

**Causes and checks:**

1. **No outbound HTTPS access** — The host cannot reach `ssh-api.k9.io:443`. Test with:
   ```bash
   curl -v https://ssh-api.k9.io/
   ```
   If this fails, check your firewall or proxy settings.

2. **DNS resolution failure** — The host cannot resolve `ssh-api.k9.io`. Test with:
   ```bash
   host ssh-api.k9.io
   ```

3. **Timeout too short** — The default `connection_timeout` is 5 seconds. If your network has high latency, increase it:
   ```yaml
   system:
     connection_timeout: 15
   ```

4. **Proxy required** — If outbound traffic must go through an HTTP proxy, set the standard Go proxy environment variables for the `key9` user:
   ```bash
   # /etc/environment or in the key9 user's profile
   HTTPS_PROXY=http://proxy.example.com:3128
   ```

---

### Non-200 status in syslog

**Symptom:**
```
k9-ssh: unexpected status from API: 401
```

**Cause by status code:**

| Status | Meaning |
|--------|---------|
| 401 | Invalid or missing API key / company UUID |
| 403 | API key does not have permission for this operation |
| 404 | Endpoint URL is wrong |
| 429 | Rate limited |
| 5xx | Key9 API server error |

**Fix for 401/403:** Verify `api_key` and `company_uuid` in `k9.yaml` exactly match the values shown in the Key9 dashboard. Credentials are combined as `company_uuid:api_key` in the request header.

---

### `config file not found` or `error loading config`

**Symptom:**
```
k9-ssh: fatal: open /opt/k9/etc/k9.yaml: no such file or directory
```

**Fix:**
```bash
# Verify the file exists
ls -la /opt/k9/etc/k9.yaml

# If using a non-default path, pass it explicitly in sshd_config:
AuthorizedKeysCommand /opt/k9/bin/k9-ssh --user=%u --config=/etc/k9/k9.yaml
```

---

### SSH login still requires password after configuration

**Symptom:** After configuring `AuthorizedKeysCommand`, SSH logins still prompt for a password.

**Checks:**

1. **sshd was not restarted** — Changes to `sshd_config` require a reload:
   ```bash
   sudo systemctl restart sshd     # Linux systemd
   sudo service sshd restart       # FreeBSD / OpenBSD
   ```

2. **sshd_config syntax error** — A syntax error prevents sshd from loading the new configuration. Check:
   ```bash
   sudo sshd -t
   ```
   Fix any errors reported before restarting.

3. **`PubkeyAuthentication` is disabled** — Ensure it is enabled:
   ```
   PubkeyAuthentication yes
   ```

4. **k9-ssh not executable** — Verify:
   ```bash
   ls -la /opt/k9/bin/k9-ssh
   # Should be: -rwxr-xr-x root root
   ```

5. **sshd cannot execute the command** — Some SELinux / AppArmor policies prevent sshd from running external binaries. Check `audit.log` or `dmesg` for denial messages.

---

### Keys returned but login still fails

**Symptom:** k9-ssh outputs valid-looking keys, but SSH login still fails.

**Checks:**

1. **Client is not using the matching private key** — The public key returned by k9-ssh must correspond to the private key the SSH client is offering. Verify by running the login with verbose output:
   ```bash
   ssh -vvv user@host
   ```
   Look for `Offering public key:` lines and compare fingerprints.

2. **`AuthorizedKeysFile none` not set, conflicting local file** — If the user has a local `~/.ssh/authorized_keys` with conflicting entries and `StrictModes yes`, sshd may reject it before consulting Key9. Check permissions on the local file.

3. **Key type not accepted by sshd** — Older sshd versions may reject newer key types. Check `sshd_config` for `PubkeyAcceptedKeyTypes` or `PubkeyAcceptedAlgorithms` restrictions.

---

## Enabling More Verbose Logging

k9-ssh always logs to syslog. To see real-time output while testing:

```bash
# Linux
tail -f /var/log/auth.log | grep k9-ssh

# macOS
log stream --predicate 'senderImagePath contains "k9-ssh"'

# OpenBSD / FreeBSD
tail -f /var/log/authlog | grep k9-ssh
```

To trace the full sshd session (useful for comparing what sshd receives):

```bash
sudo sshd -d -p 2222    # Start a debug sshd on alternate port
ssh -p 2222 user@localhost
```

---

## Still Stuck?

Contact Key9 support at [support@k9.io](mailto:support@k9.io) and include:

- The k9-ssh version (`/opt/k9/bin/k9-ssh --version`)
- Relevant syslog lines from around the time of the failed login
- The OS and OpenSSH version (`ssh -V`)
- Output of `sudo -u key9 /opt/k9/bin/k9-ssh --user=<username>` (redact any sensitive key material)
