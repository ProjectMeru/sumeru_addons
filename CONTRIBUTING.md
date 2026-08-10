# Contributing to sumeru_addons

Thanks for helping improve Sumeru’s standard business apps. This guide covers where to put changes, how to develop locally, and what we expect in pull requests.

By participating, you agree to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Where to put your work

Use the right repository so core and standard apps stay pullable for everyone:

| Change type | Repository | Notes |
| ----------- | ---------- | ----- |
| Engine, ORM, server, web shell, kernel addons (`base`, `mail`, …) | **`sumeru`** | Prefer `sumeru/core/sdk` from addon Go code; avoid new direct imports of `sumeru/core/orm` |
| Shared business apps (CRM, Sales, Accounting, …) | **`sumeru_addons`** (this repo) | Depends only on `sumeru`; folder name = technical module name |
| Customer-specific modules, branding, local runner | **`sumeru_custom_addons`** | Keep custom code under `addons/`; do not fork this repo for one-off features |

Most application teams **pull** `sumeru` and `sumeru_addons` and develop only in `sumeru_custom_addons`. Do not edit this tree for customer deployments — customize in the custom workspace so upstream pulls stay conflict-free.

## Development setup

1. Clone the three siblings (see [README.md](README.md#quick-start)).
2. Create a PostgreSQL database and configure `sumeru_custom_addons/sumeru.conf` with:

   ```ini
   addons_path = ../sumeru/addons,../sumeru_addons,./addons
   ```

3. From `sumeru_custom_addons`:

   ```bash
   make replace-sumeru
   make replace-sumeru-addons
   make generate
   make run
   ```

4. After pulling core or these addons:

   ```bash
   cd ../sumeru && git pull
   cd ../sumeru_addons && git pull
   cd ../sumeru_custom_addons && make generate
   ```

**Do not run the HTTP server from this repository.** Use `sumeru_custom_addons` so blank-imports and `addons_path` cover all three tiers.

## Adding or changing a standard module

- New apps are **direct children** of this repo (`<technical_name>/manifest.json`, optional `init.go`, `models/`, `views/`, `security/`, `data/`).
- Technical names must match `^[a-z][a-z0-9_]*$` and equal the folder name.
- This Go module must depend only on `sumeru` (plus other packages already in `go.mod`). Do not import `sumeru_custom_addons`.
- Prefer `sumeru/core/sdk` for model registration; use `event.Subscribe` for cross-module automation when appropriate.
- After structural changes, regenerate imports and install/update from the custom workspace:

  ```bash
  cd ../sumeru_custom_addons
  make generate
  go run . -- -c sumeru.conf -i your_module --stop-after-init
  # or: go run . -- -c sumeru.conf -u your_module --stop-after-init
  ```

## Testing

From this module root (and via a running custom workspace when exercising install/XML):

```bash
go build ./...
```

When behavior depends on the engine, also run tests from the `sumeru` checkout (`go test ./...`) as described in that repo’s contributing guide.

## Pull requests

- Keep diffs focused; one concern per PR when practical.
- Match existing naming and layout (addon folder = technical name; `manifest.json` + XML under `<sumeru>`).
- Do **not** commit local secrets, `sumeru.conf` with real passwords, or credentials.
- Do not commit generated custom-workspace `addonimports/` into this repo.
- Describe *why* the change is needed and how you verified it (commands, modules installed).
- For security-sensitive findings, follow [SECURITY.md](SECURITY.md) instead of a public PR discussion of exploits.

## Questions

Open a GitHub issue on [ProjectMeru/sumeru_addons](https://github.com/ProjectMeru/sumeru_addons) for design or bug discussion about these apps. Engine questions belong on [ProjectMeru/sumeru](https://github.com/ProjectMeru/sumeru). For vulnerabilities, use the private process in [SECURITY.md](SECURITY.md).
