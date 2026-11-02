# Contributing

## Documentation is part of the change

The author of a behavior change updates its documentation in the same PR. The
reviewer checks the documentation against the final implementation and evidence.
An exception needs a specific explanation in the PR; a checked box alone is not
evidence that documentation is current.

| Change | Documentation to review and update |
| --- | --- |
| Scope, support matrix, architecture or ownership | `docs/design.md`; update the README when the public summary changes |
| Installation, accounts, permissions, branding or app selection | Relevant design section and acceptance checks in `docs/validation.md` |
| DNS, TLS, proxy support or network exposure | Deployment profile and network validation scenarios |
| Authentication, secrets or administrative recovery | Administration design and recovery validation scenarios |
| Update, backup or restore behavior | Lifecycle design, recovery checks and the operator instructions introduced with the implementation |
| Upstream AIO/app version or integration change | Compatibility observations, source references and version-specific validation evidence |
| Bug fix | Reproduction, expected outcome and regression evidence; distinguish historical failures from current failures |
| Release | Supported platforms, exact tested versions, limitations, install/upgrade/recovery instructions and release notes |

Keep each detailed rule in one authoritative section and link to it from
summaries. Before merging a changed decision, reconcile related sections rather
than appending a contradictory note.

## Status and evidence

Use these distinctions consistently:

- **Approved requirement:** agreed behavior; implementation may still be absent.
- **Upstream observation:** supported by a linked upstream document or source,
  with a version or observation date; it is not a Modding runtime test.
- **Verified behavior:** supported by a recorded Modding run, including versions,
  environment and results.
- **Pending/blocked:** incomplete proof, with the remaining step or blocker named.

Verify upstream assumptions when changing the integration or preparing a release.
Record the source version and observation date. This is a contributor workflow;
no automatic recurring documentation service is configured.

Keep operator-facing UI and future operator instructions in Spanish appropriate
to Chile, using formal usted. Never include real credentials, staff rosters or
patient information in documentation, fixtures or validation artifacts. Use
synthetic identities and redact sensitive diagnostic output.

## Review checklist

For the implemented wizard, run `go test -race ./...`, `go vet ./...`,
`python3 scripts/check_sources.py` and `python3 scripts/check_docs.py`. For UI or
request-flow changes, build the executable and run the browser smoke test from
[the wizard guide](docs/wizard.md). The GitHub workflow runs those checks on PRs.

1. Compare the final change with the accepted design and identify affected docs.
2. Update those sections and remove stale commands, claims and references.
3. Check relative links, formatting and consistency between the README and docs.
4. For runtime behavior, record relevant results using the evidence fields in
   [the validation plan](docs/validation.md). Documentation-only changes do not
   establish runtime compatibility.
5. Summarize documentation changes and verification in the PR.

Before the first runnable release, add a Spanish operator guide with tested
installation, configuration, upgrade, backup and restore procedures. Until then,
describe requirements without presenting hypothetical commands as working tools.
