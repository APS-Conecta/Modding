# Validation plan and evidence

## Current status

As of 2026-11-01, the repository contains design documentation only. **No Modding
runtime checks have passed yet.** Upstream documentation/source inspection is
evidence about upstream behavior, not evidence of a working Modding installation.

The user reports that AIO already works; that is the accepted project baseline,
not a recorded Modding validation run. The first milestone is DEIS establishment
selection, editable configuration, persistence and a provisioning preview in the
Modding wizard, using the existing `gestion` visual identity. A separate proof
that upstream AIO itself works is not a prerequisite for developing that milestone.

The following checks remain release requirements, including Modding's AIO
automation and end-to-end installation on disposable servers. Detailed fixtures
and commands must be added alongside the implementation.

## Acceptance matrix

| ID | Scenario | Required outcome | Status |
| --- | --- | --- | --- |
| V01 | Fresh Ubuntu Server 24.04 LTS and Debian 13, each x86-64 | Shell entry point completes the selected suite installation on both | Pending |
| V02 | Unsupported platform, occupied ports, insufficient storage or conflicting installation | Actionable preflight result before installation changes | Pending |
| V03 | Valid release, altered payload, invalid signature and download interruption | Install only verified payloads; fail safely and support retry | Pending |
| V04 | Initial IP wizard and later IP entry to suite | Protected setup credential exchange; entry leads to intended hostname | Pending |
| V05 | LAN/VPN clients, container DNS, DNS certificate issuance and renewal | Trusted hostname TLS and correct internal reachability without public suite exposure | Pending |
| V06 | Automatic initial AIO setup | Complete without manual AIO UI actions, using unchanged official images; record tested adapter/version | Pending |
| V07 | Expired setup session, slow progressing pulls and ambiguous startup result | Bounded recovery, meaningful progress, no false success or blind replay | Pending |
| V08 | DEIS-selected establishment configuration and roster import | Preview matches the selected catalog identity and created groups, folders, named users and permissions | Pending |
| V09 | Rerun after direct administrator edits | Preserve edits and passwords; no duplicate users or deletion from absent CSV rows | Pending |
| V10 | Working and failing SMTP paths | Invitations work; failed delivery does not become completed activation | Pending |
| V11 | Included app catalog and selected optional apps | Required functions available from approved sources; failures remain visible and block completion as specified | Pending |
| V12 | Upstream IntraVox and generated welcome content | Content imports and is readable under intended service identities without exposing secrets | Pending |
| V13 | Office editing, Talk signaling and test call from client browsers | Working application flows; loading the login page alone is insufficient | Pending |
| V14 | Power/process interruption during image pull and user provisioning | Resume from verified actual state without wiping data or duplicating users | Pending |
| V15 | Privileged operation authorization, rejected inputs and local admin recovery | Narrow worker boundary; recovery expires, invalidates sessions and keeps credentials out of logs | Pending |
| V16 | Modding service upgrade with existing configuration/state | Verified release, prior-state backup and successful migration; frontend and CLI agree | Pending |
| V17 | Approved AIO update window | Pre-backup, compatibility checks, actual-version inventory and post-update checks; report image-set mismatch | Pending |
| V18 | Off-host backup and replacement-host restore | Restore AIO and Modding together with separately retained recovery material | Pending |
| V19 | Evaluation versus production completion | Backup deferral clearly visible; production readiness requires backup and restore evidence | Pending |
| V20 | Operator documentation walkthrough | Spanish instructions match actual commands, URLs, recovery and limitations | Pending |
| V21 | Two different synthetic DEIS catalog records using the same build; missing/unknown selection and catalog update | Persist and preview the correct DEIS ID, identity and catalog version; no preselected establishment or cross-installation leakage; missing/unknown selection blocks provisioning without a manual fallback; catalog update preserves an existing confirmed selection for review | Pending |
| V22 | Existing `gestion` custom theme on suite and Modding UI, including restart and update | Record source revision; verify tokens, fonts, logos, assets, desktop/mobile appearance and relevant app/guest views against the existing theme; establishment labels reflect DEIS selection; official core/image files remain unmodified and styling persists | Pending |

The required theme source is
[`gestion/themes/apsconecta`](https://github.com/APS-Conecta/gestion/tree/a09641015d1bbc7e999129795ec60c1aa96f2e54/themes/apsconecta),
including its [brand mapping](https://github.com/APS-Conecta/gestion/blob/a09641015d1bbc7e999129795ec60c1aa96f2e54/themes/apsconecta/MAPEO.md)
and shared tokens. Inspect the DEIS integration in
[`scripts/deis.py`](https://github.com/APS-Conecta/gestion/blob/a09641015d1bbc7e999129795ec60c1aa96f2e54/scripts/deis.py)
and its callers when implementing selection; identify the authoritative catalog
and record its version rather than assuming a historical CSV is current. These
references identify reuse sources, not a completed port or verified integration.

Exact hardware requirements, DNS-provider support, library versions, credential
storage, worker IPC and diagnostic retention must be documented and checked when
implemented. These are implementation proofs, not silently established facts.

## Regression sources from gestion

Issue states below were observed on 2026-11-01. Recheck them when implementing a
related fix; the table is a dated inspection, not a live issue dashboard. These
are regression candidates, not a claim that every issue affects Modding.

| Source | Observed state | Regression requirement |
| --- | --- | --- |
| [#264](https://github.com/APS-Conecta/gestion/issues/264) | Open | Test Talk signaling/calls with normal client trust; a browser certificate exception on the suite page is insufficient |
| [#247](https://github.com/APS-Conecta/gestion/issues/247) | Open | Test welcome-content import with restrictive creation permissions and the actual runtime user |
| [#173](https://github.com/APS-Conecta/gestion/issues/173) | Open | Upgrade the installed Modding service, not just apps described by an old on-disk manifest |
| [#193](https://github.com/APS-Conecta/gestion/issues/193) | Closed | Verify completion, active download progress, expired sessions and ambiguous results |
| [#243](https://github.com/APS-Conecta/gestion/issues/243) | Closed | Exercise image discovery/pulls with supported OCI manifests |
| [#188](https://github.com/APS-Conecta/gestion/issues/188), [#192](https://github.com/APS-Conecta/gestion/issues/192), [#194](https://github.com/APS-Conecta/gestion/issues/194) | Closed | Validate host/storage/DNS assumptions; report unsupported mandatory-proxy environments explicitly |

The [#175 remediation umbrella](https://github.com/APS-Conecta/gestion/issues/175)
was still open, while multiple child fixes were closed. The
[bug ledger](https://github.com/APS-Conecta/gestion/blob/main/BUGS.md) includes both
historical failures and corrections. A fix in the fork is not evidence that the
same fix is present in upstream AIO.

## Upstream observations and proof limits

Sources were inspected on 2026-11-01. Prefer release-pinned references for
implementation comparisons; mutable documentation must be rechecked for the
release being tested.

| Observation | Evidence and consequence |
| --- | --- |
| IP wizard access differs from an IP-only suite URL | [LAN guide](https://github.com/nextcloud/all-in-one/blob/v14.4.3/local-instance.md) and [domain validation](https://github.com/nextcloud/all-in-one/blob/v14.4.3/php/src/Data/ConfigurationManager.php): use the accepted hostname route |
| LAN hostname TLS through DNS verification is documented | [Reverse-proxy guide](https://github.com/nextcloud/all-in-one/blob/v14.4.3/reverse-proxy.md): prove DNS, issuance, renewal and service reachability |
| First-setup automation has no verified stable public API | [Routes](https://github.com/nextcloud/all-in-one/blob/v14.4.3/php/public/index.php) and [external lifecycle documentation](https://github.com/nextcloud/all-in-one/blob/v14.4.3/readme.md#how-to-stopstartupdate-containers-or-trigger-the-daily-backup-from-a-script-externally): internal setup adapter remains a release gate |
| EuroOffice is a stock fresh-install default in v14.4.3 | [Configuration manager](https://github.com/nextcloud/all-in-one/blob/v14.4.3/php/src/Data/ConfigurationManager.php): verify selected-release behavior and browser editing |
| AIO startup owns some integrated app settings and updates | [Nextcloud entrypoint](https://github.com/nextcloud/all-in-one/blob/v14.4.3/Containers/nextcloud/entrypoint.sh): map ownership before adding reconciliation |
| Image updates pull by moving tag | [Docker actions](https://github.com/nextcloud/all-in-one/blob/v14.4.3/php/src/Docker/DockerActionManager.php): preflight checks cannot guarantee a frozen image set |
| Docker daemon proxies do not cover all container networking | [Docker proxy docs](https://docs.docker.com/engine/daemon/proxy/) and [AIO discussion](https://github.com/nextcloud/all-in-one/discussions/5634): authenticated-proxy deployments remain deferred |
| Email-based account activation is supported | [Provisioning API](https://docs.nextcloud.com/server/stable/admin_manual/configuration_user/instruction_set_for_users.html#add-a-new-user): verify SMTP and invitation delivery |
| Native temporary passwords do not establish forced first-login changes | [Password-policy request](https://github.com/nextcloud/password_policy/issues/14): keep the email-activation design |
| Standard AIO restore omits additional backup directories | [Backup UI](https://github.com/nextcloud/all-in-one/blob/v14.4.3/php/templates/containers.twig): implement and rehearse the Modding-state recovery step |
| Nextcloud downgrade is unsupported | [Upgrade documentation](https://docs.nextcloud.com/server/stable/admin_manual/maintenance/upgrade.html) and [restore documentation](https://docs.nextcloud.com/server/stable/admin_manual/maintenance/restore.html): recover consistent data/configuration/software, not only an old image |

## Recording a validation run

Record the following with each result:

- Date, source commit and Modding release/build identifier.
- OS, architecture, resource allocation and relevant network/DNS profile.
- AIO mastercontainer and child image versions/digests, Nextcloud version and
  actual app identifiers/versions/sources.
- Covered acceptance IDs, steps or reproducible automation, expected and actual
  outcomes, and pass/fail/blocked status.
- Redacted logs and browser evidence for UI, Office and Talk checks.
- Backup/restore evidence, known limitations and remaining blockers.

Use synthetic establishments and accounts. Preserve evidence outside disposable
test hosts. A test on one OS or image/app combination does not certify another.
Documentation checks may pass while every runtime check remains pending.
