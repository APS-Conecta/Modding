# Repository guidance

- Before changing installation, provisioning, app selection, networking,
  authentication, updates or recovery, read [docs/design.md](docs/design.md).
  Preserve its accepted boundaries; surface incompatible requirements explicitly.
- When implementing or changing a covered behavior, read
  [docs/validation.md](docs/validation.md) and link relevant validation evidence.
  Pending checks become passed only with a recorded run against identified versions.
- For every change, follow [CONTRIBUTING.md](CONTRIBUTING.md). Update affected
  documentation in the same PR, or explain specifically why no update is needed.
- Keep planned behavior distinct from implemented behavior. Publish runnable
  installation commands only after the referenced release and commands exist.
- Keep upstream images unmodified. Isolate internal AIO setup integration and
  document the exact tested upstream version and any unsupported behavior.
