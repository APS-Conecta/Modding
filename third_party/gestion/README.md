# Imported Gestión sources

`sources.json` records the exact upstream revision, paths and SHA-256 hashes for
the imported assets and public establishment catalog. `scripts/check_sources.py`
verifies the local bytes. Imported files retain their original contents.
`.gitattributes` preserves the source catalog's CRLF endings and the original
whitespace in SVG assets and font notices; authored source uses LF endings.

The theme assets and brand mapping come from `APS-Conecta/gestion` under
AGPL-3.0-or-later; the repository carries its AGPL license text in `LICENSE`.
Fraunces and Nunito Sans are distributed under SIL OFL 1.1; the family notices
are included here, obtained from the corresponding `google/fonts/ofl` directories.

The embedded catalog is `sites/establecimientos-deis-2026-07-23.csv`, the snapshot
actually loaded by `scripts/deis.py` and the Provisionador at the recorded source
revision. It contains 2,655 APS establishments and MINSAL/DEIS Geoportal coordinates.
The larger November master file uses a different schema and is not the selector's
runtime input. This import deliberately preserves the runtime snapshot and its
date, rather than relabeling it as a newer dataset. Upstream attribution and data
licensing context: [Gestión licensing documentation](https://aps-conecta.github.io/documentation/desarrollo/licencias.html).

The wizard embeds original tokens, regular font files, logos, favicon and background
asset. Its page layout is new; the complete Nextcloud server theme has not yet
been installed or validated against upstream Nextcloud. `MAPEO.md` describes the
original theme and may link to files that remain in the source repository.

To update an import, inspect the upstream revision, replace the selected bytes,
update `sources.json`, record the catalog date accurately, and run the catalog,
provenance, browser and state-migration checks. Never update a saved installation's
DEIS identity silently when changing the bundled catalog.
