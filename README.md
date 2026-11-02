# Modding

Modding is the planned APS Conecta installer and administration layer for an
unmodified Nextcloud All-in-One (AIO) installation. It will configure the suite,
create named users, apply the existing `gestion` custom theme, and install the selected applications for
one establishment per server. Modding is establishment-agnostic: the operator
selects the establishment from the DEIS catalog during setup. No particular
establishment is hard-coded or preselected in the installer.

**Status: design approved; implementation and runtime validation pending.**
This repository does not yet provide an installer or a downloadable release.
The user reports that AIO already works. The next implementation milestone is
the Modding configuration wizard, using the existing APS visual identity: select
an establishment through DEIS, edit its configuration, save it and preview the
provisioning plan. Validation will focus on Modding's integration and changes to
that working base.

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
