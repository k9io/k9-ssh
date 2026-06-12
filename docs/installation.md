# Installation

## Prerequisites

- A Key9 account with your **Company UUID** and **API Key** (available in the Key9 dashboard)
- A Key9 **machine group** configured for the hosts you are installing on
- OpenSSH server (`sshd`) installed and running
- Root or sudo access on the target host

## Download a Pre-Built Binary

Pre-built static binaries are available for all supported platforms at [https://github.com/k9io/k9-binaries](https://github.com/k9io/k9-binaries). Download the binary for your OS and architecture, verify the checksum, and install it.

```bash
# Example: Linux amd64
curl -LO https://github.com/k9io/k9-binaries/raw/main/k9-ssh/linux/k9-ssh.amd64.gz
curl -LO https://github.com/k9io/k9-binaries/raw/main/k9-ssh/linux/k9-ssh.amd64.gz-sha256.txt

sha256sum -c k9-ssh.amd64.gz-sha256.txt
gunzip k9-ssh.amd64.gz
chmod 755 k9-ssh
```

Replace `linux` and `amd64` with your platform and architecture as needed.

## Create the `key9` System User

k9-ssh must run as a dedicated, unprivileged system user. Create the `key9` user and group before installing:

**Linux (Debian / Ubuntu)**

```bash
sudo addgroup --quiet --system key9
sudo adduser --quiet --system --no-create-home --disabled-password \
    --disabled-login --shell /bin/false --ingroup key9 --home / key9
```

**Linux (RHEL / Fedora / AlmaLinux)**

```bash
sudo groupadd --system key9
sudo useradd --system --no-create-home --shell /sbin/nologin \
    --gid key9 --home-dir / key9
```

**FreeBSD**

```bash
sudo pw groupadd key9 -g 900
sudo pw useradd key9 -u 900 -g key9 -d / -s /sbin/nologin \
    -c "Key9 SSH agent" -w no
```

**OpenBSD**

```bash
sudo groupadd -g 900 key9
sudo useradd -u 900 -g key9 -d / -s /sbin/nologin \
    -c "Key9 SSH agent" key9
```

## Install the Binary and Configuration

```bash
# Create directory structure
sudo mkdir -p /opt/k9/bin /opt/k9/etc

# Install binary
sudo cp k9-ssh /opt/k9/bin/k9-ssh
sudo chown root:root /opt/k9/bin/k9-ssh
sudo chmod 755 /opt/k9/bin/k9-ssh

# Install example configuration
sudo cp etc/k9.yaml /opt/k9/etc/k9.yaml
sudo chown root:key9 /opt/k9/etc/k9.yaml
sudo chmod 640 /opt/k9/etc/k9.yaml
```

## Configure k9-ssh

Edit `/opt/k9/etc/k9.yaml` and fill in your Key9 credentials and machine group:

```bash
sudo nano /opt/k9/etc/k9.yaml
```

See [Configuration](configuration.md) for a full reference.

## Configure sshd

Add the following lines to `/etc/ssh/sshd_config` (path may differ by OS — see the table below):

```
AuthorizedKeysCommand /opt/k9/bin/k9-ssh --user=%u
AuthorizedKeysCommandUser key9
```

If your OpenSSH version is 9.4 or later you can also pass the client address, which Key9 can use for additional policy checks:

```
AuthorizedKeysCommand /opt/k9/bin/k9-ssh --user=%u --remote=%C
AuthorizedKeysCommandUser key9
```

**sshd_config locations by OS**

| OS | Default path |
|----|-------------|
| Linux (most) | `/etc/ssh/sshd_config` |
| FreeBSD | `/etc/ssh/sshd_config` |
| OpenBSD | `/etc/ssh/sshd_config` |
| NetBSD | `/etc/ssh/sshd_config` |
| Solaris | `/etc/ssh/sshd_config` |

## Restart sshd

**Linux (systemd)**

```bash
sudo systemctl restart ssh    # Debian/Ubuntu
sudo systemctl restart sshd   # RHEL/Fedora
```

**FreeBSD / OpenBSD / NetBSD**

```bash
sudo service sshd restart
```

## Verify the Installation

Test k9-ssh directly before relying on it for login. Run it as the `key9` user with a known Key9 username:

```bash
sudo -u key9 /opt/k9/bin/k9-ssh --user=<key9-username>
```

If the user has keys registered in Key9 you should see one or more `ssh-*` public key lines printed to stdout. A blank response means the user has no keys or does not belong to the configured machine group.

> **Important:** Keep your existing SSH session open while testing. Only restart sshd and test a new login after confirming k9-ssh returns keys correctly.
