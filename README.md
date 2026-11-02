# Modding

Modding is the APS Conecta installer and administration layer being built for an
unmodified Nextcloud All-in-One (AIO) installation. It will configure the suite,
create named users, apply the existing `gestion` custom theme, and install the selected applications for
one establishment per server. Modding is establishment-agnostic: the operator
selects the establishment from the DEIS catalog during setup. No particular
establishment is hard-coded or preselected in the installer.

**Status: configuration wizard implemented; suite installation and provisioning pending.**
The Go wizard provides DEIS search/selection, editable organizational configuration,
SQLite persistence, a provisioning preview, confirmation and JSON export. It uses
the original Gestión brand tokens, fonts and logos. AIO is the working upstream
base; this build does not yet install or modify it, or apply the suite theme.

To run the wizard locally with the Go version in [go.mod](go.mod):

```sh
go build -o bin/modding ./cmd/modding
./bin/modding --state-dir .modding
```

Open `http://127.0.0.1:8787` and enter the one-use code printed in the terminal.
See the [Spanish wizard guide](docs/wizard.md) for network access with TLS,
catalog updates and verification. There is no production installer or published
downloadable release yet.

The agreed delivery method is an HTTPS shell bootstrap, usable through
`curl … | sh`, which downloads and verifies a versioned Modding release. There
will be no `.deb` package. Installation targets fresh, dedicated Ubuntu Server
24.04 LTS and Debian 13 hosts on x86-64, with internet access for downloads.

The setup interface will open at the server IP. An IP entry page will direct
users to the finished suite's configured hostname, which resolves inside the
establishment's LAN or VPN and uses a publicly trusted certificate. Permanent
IP-only Nextcloud URLs are outside the agreed deployment profile.

## Project documentation

- [Design](docs/design.md): agreed scope, architecture, ownership and operations.
- [Validation](docs/validation.md): technical proof requirements, regression
  scenarios and release evidence.
- [Contributing](CONTRIBUTING.md): documentation upkeep and change requirements.
- [Wizard guide](docs/wizard.md): run and use the implemented configuration milestone.

Behavior changes must update the corresponding documentation in the same PR.
Documentation distinguishes approved requirements, upstream observations and
capabilities actually verified in Modding.

## Upstream relationship

The project draws on [APS Conecta Gestión](https://github.com/APS-Conecta/gestion)
while using official [Nextcloud AIO](https://github.com/nextcloud/all-in-one)
images and original upstream third-party apps. APS-specific applications remain
separate extensions. Nextcloud/AIO source and container files are not patched.
The design and validation documents identify integration limits that still need
proof; this repository does not claim affiliation with Nextcloud GmbH.
