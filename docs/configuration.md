# Configuration

k9-ssh is configured through a single YAML file. The default path is `/opt/k9/etc/k9.yaml`. You can specify an alternate path with the `--config` flag.

## Minimal Configuration

```yaml
system:
  machine_group: "production"
  run_as: "key9"

authentication:
  api_key: "YOUR_API_KEY"
  company_uuid: "YOUR_COMPANY_UUID"

urls:
  query_ssh_keys: "https://ssh-api.k9.io/api/v1/ssh/query/"
```

## Full Configuration Reference

```yaml
system:
  # Required. The Key9 machine group (or comma-delimited groups) this host
  # belongs to. Keys are only returned for users that have access to at least
  # one of these groups.
  machine_group: "production"

  # Required. The OS user that k9-ssh must run as. k9-ssh will exit if the
  # current process user does not match this value. Must match the user set
  # in sshd_config's AuthorizedKeysCommandUser directive.
  run_as: "key9"

  # Optional. Timeout in seconds for API requests. Defaults to 5.
  connection_timeout: 5

authentication:
  # Required. Your Key9 company API key.
  api_key: "YOUR_API_KEY"

  # Required. Your Key9 company UUID.
  company_uuid: "YOUR_COMPANY_UUID"

urls:
  # Required. The Key9 SSH key query endpoint. Do not change unless instructed
  # by Key9 support.
  query_ssh_keys: "https://ssh-api.k9.io/api/v1/ssh/query/"

  # The following URL fields are used by the Key9 NSS library (libnss-k9)
  # and are not required for k9-ssh alone. Include them if you are also
  # deploying the NSS library on this host.
  query_all_users:       "https://ssh-api.k9.io/api/v1/query/k9/all_users"
  query_group_name:      "https://ssh-api.k9.io/api/v1/query/group/name"
  query_group_gid:       "https://ssh-api.k9.io/api/v1/query/group/gid"
  query_group_id:        "https://ssh-api.k9.io/api/v1/query/group/id"
  query_shadow_username: "https://ssh-api.k9.io/api/v1/query/shadow/username"
  query_passwd_username: "https://ssh-api.k9.io/api/v1/query/passwd/username"
  query_passwd_uid:      "https://ssh-api.k9.io/api/v1/query/passwd/uid"
  query_passwd_id:       "https://ssh-api.k9.io/api/v1/query/passwd/id"

tail:
  # Used by the k9-tail log shipping component (not required for k9-ssh).
  tail_file:          "/var/log/auth.log"
  waldo_file:         "/opt/k9/cache/auth.waldo"
  client_logging_url: "https://client-logging.k9-api.io/client-logging/api/v1/post"
```

## Field Reference

### `system` section

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `machine_group` | Yes | — | Comma-delimited Key9 machine group(s) this host belongs to, e.g. `"dev,bastion"`. Keys are scoped to users who have access to one or more of these groups. |
| `run_as` | Yes | — | OS username under which k9-ssh must run. k9-ssh exits immediately if the current user does not match. |
| `connection_timeout` | No | `5` | HTTP timeout (seconds) for Key9 API requests. |

### `authentication` section

| Field | Required | Description |
|-------|----------|-------------|
| `api_key` | Yes | Your Key9 company API key from the Key9 dashboard. |
| `company_uuid` | Yes | Your Key9 company UUID from the Key9 dashboard. |

These two values are combined into a single `API_KEY` header (`company_uuid:api_key`) on every API request.

### `urls` section

| Field | Required | Description |
|-------|----------|-------------|
| `query_ssh_keys` | Yes | Base URL for SSH key lookups. k9-ssh appends `{username}/{machine_group}` to this URL when querying. Use `https://` for the Key9 cloud API, or `http://` when pointing at a local [k9-proxy](k9-proxy.md) instance. |
| All others | No | Used by the Key9 NSS library; leave them at their defaults or omit entirely if not using the NSS library. |

## Machine Groups

Machine groups let you segment access. A user in Key9 must be granted access to at least one of the groups listed in `machine_group` for their keys to be returned by k9-ssh on that host.

You can assign multiple groups to a single host using a comma-delimited list:

```yaml
system:
  machine_group: "bastion,production,monitoring"
```

A single Key9 user can then be granted access to just `bastion` to reach only jump hosts, or to `production` to reach application servers, without any reconfiguration on the hosts themselves.

## File Permissions

The configuration file contains your API credentials and should be readable only by root and the `key9` group:

```bash
sudo chown root:key9 /opt/k9/etc/k9.yaml
sudo chmod 640 /opt/k9/etc/k9.yaml
```

## Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--user` | (required) | The OS username being authenticated. sshd passes this automatically via `%u`. |
| `--remote` | (none) | Client connection string. sshd passes this automatically via `%C` (OpenSSH 9.4+). |
| `--config` | `/opt/k9/etc/k9.yaml` | Path to the YAML configuration file. |
