
Join the Key9 Slack channel
---------------------------

[![Slack](./images/slack.png)](https://key9identity.slack.com/)


What is they Key9-SSH?
------------------------

“k9-ssh” is a small program used to retrieve public keys from the Key9 SSH API. Key9 allows for easy addition and removal of SSH keys for authentication. This program works with OpenSSH via the <b>AuthorizedKeysCommand</b> command.  For example,  this would be added to your /etc/sshd/sshd_config : 

<pre>
AuthorizedKeysCommand /opt/k9/bin/k9-ssh --user=%u
AuthorizedKeysCommandUser key9
</pre>
 
Note:  The option <b>--remote="%C"</b> can be added if you are using OpenSSH version 9.4/9.4p1 (2023-08-10) or higher

This program relies on you having a company UUID and API key registered with Key9.


Use cases:
----------

Key9 is a provider of "Identity and Access Management" that offers a completely "passwordless" solution. The concept behind Key9 is that by not storing passwords, there is nothing for hackers to steal.

Part of the Key9 services involves managing user access to the operating system (Linux, OpenBSD, NetBSD, etc) via Secure Shell (SSH) using public key cryptography.

This repo contains 'k9-ssh', which serves as a bridge between OpenSSH and the Key9 SSH API.


What software uses they Key9 SSH?
---------------------------------

OpenSSH


Building and installing the Key9 SSH 
------------------------------------

Make sure you have Golang installed! 

Add the "key9" user and group. k9-ssh runs as this user to protect the security of the system. 

<pre>
$ sudo addgroup --quiet --system key9
$ sudo adduser --quiet --system --no-create-home --disabled-password --disabled-login --shell /bin/false --ingroup key9 --home / key9 
</pre>

Compiling and installing k9-ssh.  Make sure you have Golang installed!

<pre>
$ go mod init k9-ssh
$ go mod tidy
$ go build
$ sudo mkdir -p /opt/k9/etc /opt/k9/bin
$ sudo cp etc/k9.yaml /opt/k9/etc
$ sudo cp k9-ssh /opt/k9/bin
$ sudo nano /opt/k9/etc/k9.yaml    # Modify you company UUID, API Key and assign a "group" to the machine.
$ sudo nano /etc/sshd/sshd_config  # Added the AuthorizedKeysCommand/AuthorizedKeysCommandUser specified above.
</pre>

Prebuild Key9 SSH binaries
--------------------------

If you are unable to access a Golang compiler, you can download pre-built/pre-compiled binaries. These binaries are available for various architectures (i386, amd64, arm64, etc) and multiple operating systems (Linux, Solaris, NetBSD, etc).

You can find those binaries at: https://github.com/k9io/k9-binaries/tree/main/k9-ssh

You will need a copy of the 'k9-ssh' configuation file.  That is located at: 

https://github.com/k9io/k9-ssh/blob/main/etc/k9.yaml

