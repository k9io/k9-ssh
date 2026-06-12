# Using k9-proxy

**k9-proxy** is an optional proxy for the Key9 API that can be deployed in two ways:

- **Local cache** — Runs on the same host as k9-ssh, bound to the loopback interface. SSH authentication continues working even when the Key9 cloud API is unreachable because k9-proxy serves responses from its on-disk cache. TLS is not required in this mode since traffic never leaves the host.
- **Network gateway** — Runs on a central host and serves multiple machines over the network. Hosts in restricted environments that cannot reach the Key9 cloud API directly route their requests through this proxy. TLS should be enabled in this mode to protect API credentials in transit.

## How It Works

Without k9-proxy, k9-ssh calls the Key9 cloud API directly on every login attempt:

```
k9-ssh → https://ssh-api.k9.io  (Key9 cloud)
```

With k9-proxy running locally, k9-ssh calls the proxy instead. The proxy forwards the request upstream, caches the successful response to disk, and returns it to k9-ssh. If the upstream API is later unavailable, the proxy serves the cached response:

```
k9-ssh → http://127.0.0.1:8080  (k9-proxy, local)
              └→ https://ssh-api.k9.io  (Key9 cloud, when reachable)
                        ↕  cached to /opt/k9/proxy_cache on success
```

## TLS Is Not Required for Local Deployments

When k9-proxy is bound to `127.0.0.1` (loopback only), traffic never leaves the host, so TLS is unnecessary. k9-ssh can be pointed at the proxy using a plain `http://` URL. This is the recommended local-cache setup.

TLS should be enabled on k9-proxy only when it is serving multiple hosts over a network (see [Network Gateway Mode](#network-gateway-mode) below).

## Configuring k9-ssh to Use k9-proxy

Change the `query_ssh_keys` URL in `/opt/k9/etc/k9.yaml` to point at the local proxy instead of the cloud API:

```yaml
urls:
  query_ssh_keys: "http://127.0.0.1:8080/api/v1/ssh/query/"
```

If you are also using the Key9 NSS library on the same host, update the remaining URL fields to match:

```yaml
urls:
  query_ssh_keys:        "http://127.0.0.1:8080/api/v1/ssh/query/"
  query_all_users:       "http://127.0.0.1:8080/api/v1/query/k9/all_users"
  query_group_name:      "http://127.0.0.1:8080/api/v1/query/group/name"
  query_group_gid:       "http://127.0.0.1:8080/api/v1/query/group/gid"
  query_group_id:        "http://127.0.0.1:8080/api/v1/query/group/id"
  query_shadow_username: "http://127.0.0.1:8080/api/v1/query/shadow/username"
  query_passwd_username: "http://127.0.0.1:8080/api/v1/query/passwd/username"
  query_passwd_uid:      "http://127.0.0.1:8080/api/v1/query/passwd/uid"
  query_passwd_id:       "http://127.0.0.1:8080/api/v1/query/passwd/id"
```

No other changes to k9-ssh or `sshd_config` are required. k9-proxy implements the exact same API contract as the Key9 cloud, making it a transparent drop-in replacement.

## k9-proxy Configuration (Local Cache)

k9-proxy has its own YAML configuration file, typically at `/opt/k9/etc/k9-proxy.yaml`. A minimal local-cache setup with TLS disabled:

```yaml
core:
  address: "https://ssh-api.k9.io"   # Upstream Key9 cloud API
  runas: "key9"
  connection_timeout: 5

proxy:
  http_listen: "127.0.0.1:8080"      # Loopback only — no TLS needed
  http_mode: "release"
  http_tls: false
  cache_dir: "/opt/k9/proxy_cache"
```

The `cache_dir` directory must exist and be writable by the `key9` user:

```bash
sudo mkdir -p /opt/k9/proxy_cache
sudo chown key9:key9 /opt/k9/proxy_cache
sudo chmod 700 /opt/k9/proxy_cache
```

## Network Gateway Mode

If k9-proxy is running on a central host and serving multiple machines over the network, TLS should be enabled:

```yaml
proxy:
  http_listen: "0.0.0.0:443"
  http_mode: "release"
  http_tls: true
  http_cert: "/opt/k9/etc/proxy.crt"
  http_key:  "/opt/k9/etc/proxy.key"
  cache_dir: "/opt/k9/proxy_cache"
```

k9-ssh on the client hosts then points at the proxy over HTTPS:

```yaml
urls:
  query_ssh_keys: "https://proxy.internal.example.com/api/v1/ssh/query/"
```

In this mode TLS protects the API key in transit across the network.

## Cache Behavior

- Successful API responses are cached to disk, keyed by a SHA-256 hash of the request URL.
- Error responses (e.g., user not found) are **never** cached.
- When the upstream API is unreachable, k9-proxy validates the incoming API credentials against a cached credential file and returns the last known good response.
- Cache files are written with mode `0600` (readable only by the `key9` user).

This means that after the first successful login for a user, subsequent logins will succeed even during a Key9 cloud outage, as long as k9-proxy is running locally.

## Verifying the Setup

Test the full chain by querying k9-proxy directly:

```bash
# Replace <uuid>, <apikey>, <user>, and <group> with your values
curl -s -X POST \
  -H "API_KEY: <uuid>:<apikey>" \
  http://127.0.0.1:8080/api/v1/ssh/query/<user>/<group>
```

Then confirm k9-ssh reads through the proxy:

```bash
sudo -u key9 /opt/k9/bin/k9-ssh --user=<username>
```

Both commands should return the same public key lines.
