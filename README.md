
Join the Key9 Slack channel
---------------------------

[![Slack](./images/slack.png)](https://key9identity.slack.com/)


What is they Key9-SSH?
------------------------

“k9-ssh” is a small program used to retrieve public keys from the Key9 SSH API. Key9 allows for easy addition and removal of SSH keys for authentication. This program works with OpenSSH via the <b>AuthorizedKeysCommand</b> command.

<pre>
AuthorizedKeysCommand /opt/k9/bin/k9-ssh --user=%u --remote="%C"
AuthorizedKeysCommandUser key9
<pre>
 
Note:  The option <b>--remote="%C"</b> can be added if you are using OpenSSH version 9.4/9.4p1 (2023-08-10) or higher

This program relies on you having a company UUID and API key registered with Key9.


Use cases:
----------

What software uses they Key9 Proxy?
-----------------------------------

The proxy is used by k9-ssh (public key retrieval) and k9-nss (operating system NSS library)

Building and installing the Key9 Proxy
--------------------------------------

Make sure you have Golang installed! 

<pre>
$ go mod init k9-proxy
$ go mod tidy
$ go build
$ sudo mkdir -p /opt/k9/etc /opt/k9/bin
$ sudo cp etc/k9-proxy.yaml /opt/k9/etc
$ sudo cp k9-proxy /opt/k9/bin
$ sudo /opt/k9/bin/k9-proxy 	 # Run from the command line... Control C exits
$ sudo cp k9-proxy.service /etc/systemd/system
$ sudo systemctl enable k9-proxy
$ sudo systemctl start k9-proxy
</pre>

Prebuild Key9 proxy binaries
----------------------------

If you are unable to access a Golang compiler, you can download pre-built/pre-compiled binaries. These binaries are available for various architectures (i386, amd64, arm64, etc) and multiple operating systems (Linux, Solaris, NetBSD, etc).

You can find those binaries at: https://github.com/k9io/k9-binaries/tree/main/k9-proxy

You will need a copy of the 'k9-proxy' configuation file.  That is located at: 

https://github.com/k9io/k9-proxy/blob/main/etc/k9-proxy.yaml

