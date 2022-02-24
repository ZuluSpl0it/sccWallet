# WebWallet

ScPrime has released a browser based GUI (Graphical User Interface) wallet called the WebWallet, that is purely for storing your SCP and doesn't have the 'hosting' aspect that the CLI or UI software has. If you're just looking for software to just send, receive, and hold your SCP, this is a good option for you.

Online documentation available at https://docs.scpri.me/software/webwallet

## Building From Source

To build from source, Go 1.18 or above must be installed on the system. Clone the repo and run make:

```sh
git clone https://gitlab.com/scpcorp/webwallet
cd webwallet && make
```
This will install the `scp-ui` binary in your $GOPATH/bin folder (By default, this is $HOME/go/bin).

