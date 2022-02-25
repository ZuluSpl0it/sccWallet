# [![ScPrime WebWallet](https://scpri.me/imagestore/SPRho_256x256.png)](http://scpri.me)

[![Go Report Card](https://goreportcard.com/badge/gitlab.com/scpcorp/webwallet)](https://goreportcard.com/report/gitlab.com/scpcorp/webwallet)
[![License MIT](https://img.shields.io/badge/License-MIT-brightgreen.svg)](https://img.shields.io/badge/License-MIT-brightgreen.svg)

ScPrime has released a browser based GUI (Graphical User Interface) wallet called the WebWallet, that is purely for storing your SCP and doesn't have the 'hosting' aspect that the CLI or UI software has. If you're just looking for software to just send, receive, and hold your SCP, this is a good option for you.

Online documentation available at https://docs.scpri.me/software/webwallet

Building From Source
--------------------

To build from source, [Go 1.17 or above must be installed](https://golang.org/doc/install) on the system. Clone the repo and run make:

```sh
git clone https://gitlab.com/scpcorp/webwallet
cd webwallet && make
```

This will install the `scp-webwallet` binary in your $GOPATH/bin folder (By default, this is $HOME/go/bin).

