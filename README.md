# sumeru_addons

[![Go](https://img.shields.io/badge/Go-1.26.2+-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Pre-Alpha](https://img.shields.io/badge/Status-Pre--Alpha-critical?style=for-the-badge)](https://github.com/ProjectMeru/sumeru)

> [!CAUTION]
> ### 🚧 Pre-Alpha Software
>
> Sumeru and its standard addons are **pre-alpha software**. They are under active development and are **not ready for production or commercial use**.
>
> - **No production use.** Do not deploy to production or run live business workloads. Stability, security, and data integrity are not guaranteed.
> - **Not for sale.** Do not offer, resell, license, or deploy Sumeru to customers. This is not a commercial product.
> - **Evaluation only.** Use for local development, testing, and feedback at your own risk.
>
> APIs, data models, and behavior may change without notice. There is no migration guarantee and no production support.

Standard **business addons** for [Sumeru](https://github.com/ProjectMeru/sumeru) (Go module: `module sumeru_addons`).

This repository is **Tier 2** of the Sumeru stack: shared CRM, Sales, Accounting, Purchase, HR, and related apps. It depends only on **[`../sumeru`](../sumeru/)** (`replace sumeru => ../sumeru` in `go.mod`).

## Do not modify this tree for customer work

**Business and customer teams should treat this repository as pull-only.**

- Pull updates and install the modules you need.
- Put overrides, branding, and client-specific modules under **[`sumeru_custom_addons/addons/`](../sumeru_custom_addons/)**.
- Do **not** edit files here to customize a deployment. That creates merge conflicts and blocks clean upstream pulls.

Shared improvements that belong in the standard suite should land via a pull request to this repo, not by forking and diverging locally.

**Do not run the HTTP server from this tree.** Always run from **[`../sumeru_custom_addons`](../sumeru_custom_addons/)** so `addons_path` can list core, these addons, and your custom addons together.

## Architecture and module placement

Sumeru is split into three sibling repositories so you can pull the engine and standard apps without mixing in customer code.

```text
parent/
  sumeru/                 # Tier 1: engine + kernel addons (pull-only)
  sumeru_addons/          # Tier 2: this repo (pull-only for businesses)
  sumeru_custom_addons/   # Tier 3: run the server + custom addons
```

```text
sumeru_custom_addons  ──replace + make generate──►  sumeru (core)
         │                                              │
         └──replace + addons_path──────────────────────►│
         │                                              ▼
         └──make run────────────────────────────►  HTTP server
                ▲
                └── also loads  sumeru_addons (this repo)
```

| Repository | Role | Remote |
| ---------- | ---- | ------ |
| **`sumeru`** | Core engine + kernel addons (`base`, `mail`, …) | `git@github.com:ProjectMeru/sumeru.git` |
| **`sumeru_addons`** | Standard business apps (this repo) | `git@github.com:ProjectMeru/sumeru_addons.git` |
| **`sumeru_custom_addons`** | Workspace: custom addons, local INI, generated imports, process you run | `git@github.com:ProjectMeru/sumeru_custom_addons.git` |

### How core, standard, and custom modules communicate

| Mechanism | How it works |
| --------- | ------------ |
| **`addons_path`** | Comma-separated roots in `sumeru.conf`, e.g. `../sumeru/addons,../sumeru_addons,./addons`. Later roots **override** the same `manifest.name`. |
| **Go modules** | This repo replaces `sumeru => ../sumeru`. The custom workspace also replaces `sumeru_addons` and generates blank-imports so `init()` runs. |
| **Manifests** | Each app has `manifest.json` with `depends`, `data` (XML/CSV), and optional `application`. |
| **Models** | Go structs with `sdk.Model` embed; `make generate` → `sdk.MustRegister` |
| **XML / security** | Views, menus, groups, and ACLs load on install/update. |
| **Events** | Cross-module automation via `event.Subscribe` (see [`sale_crm`](sale_crm/)). |
| **Relations** | Many2One / related fields to core models such as `core.partner`, `core.user`, `core.company`. |

List **`../sumeru_addons`** on `addons_path` **before** **`./addons`** so custom modules can extend standard apps without forking them.

## Included modules

| Module | Role |
| ------ | ---- |
| `contacts` | Address book (views/security on `core.partner`) |
| `product` | Product catalog |
| `crm` | Leads / opportunities |
| `sale` | Quotations / sales orders |
| `sale_crm` | Opportunity → draft quotation (on Won) |
| `account` | Invoices, bills, COA, journal lines |
| `purchase` | RFQ / PO → vendor bill |
| `hr` | Employees, departments, jobs |

Typical install order: `contacts, product, crm, sale, sale_crm, account, purchase, hr`.

## Prerequisites

| Requirement | Notes |
| ----------- | ----- |
| [Go](https://go.dev/dl/) **1.26.2+** | See `go.mod` |
| [PostgreSQL](https://www.postgresql.org/) | Application database |
| Sibling checkouts | `sumeru` and `sumeru_custom_addons` next to this repo |

## Quick start

Clone all three repos as siblings, then configure and run from the custom workspace.

```bash
mkdir -p ~/sumeru_erp && cd ~/sumeru_erp
git clone git@github.com:ProjectMeru/sumeru.git
git clone git@github.com:ProjectMeru/sumeru_addons.git
git clone git@github.com:ProjectMeru/sumeru_custom_addons.git

# Create an empty database matching db_name in your INI, e.g.:
#   psql -c "CREATE DATABASE sumeru;"

cd sumeru_custom_addons
cp sumeru.conf.example sumeru.conf   # edit db_* , http_port, addons_path
make replace-sumeru
make replace-sumeru-addons
make generate                        # → addonimports/zimports.go
make run
```

Ensure `addons_path` includes this tree, for example:

```ini
addons_path = ../sumeru/addons,../sumeru_addons,./addons
```

Full workspace details: sibling **[`sumeru_custom_addons/README.md`](https://github.com/ProjectMeru/sumeru_custom_addons/blob/main/README.md)**.

### Day-to-day updates

```bash
cd ../sumeru && git pull
cd ../sumeru_addons && git pull
cd ../sumeru_custom_addons && make generate && make run
```

### Install / update modules

From `sumeru_custom_addons`:

```bash
go run . -- -c sumeru.conf -i contacts,product,crm,sale,sale_crm,account,purchase,hr --stop-after-init
go run . -- -c sumeru.conf -u sale --stop-after-init
go run . -- -c sumeru.conf
```

## Module layout

Each installable app is a **direct child directory** of this repo (sibling to `go.mod`), with `manifest.json` and optional `init.go` / `models/`:

```text
sumeru_addons/
  go.mod
  <technical_name>/
    manifest.json
    init.go              # optional: blank-import models; event hooks
    models/ …
    views/ …
    security/ …
    data/ …
```

Technical module names must match **`^[a-z][a-z0-9_]*$`** and equal the **folder name**.

Reference patterns in this repo:

- **Views on a core model:** [`contacts/`](contacts/) (no new models; XML + security on `core.partner`)
- **Bridge / events:** [`sale_crm/`](sale_crm/) (`event.Subscribe` when a lead is Won)

## How to create or extend a module

### Customer / business teams (recommended)

1. Keep **`sumeru`** and **`sumeru_addons`** pull-only.
2. Create modules under **`sumeru_custom_addons/addons/<technical_name>/`** with the usual layout (`manifest.json`, `init.go`, models, views, security).
3. Declare `depends` on standard apps as needed (e.g. `sale`, `crm`).
4. Blank-import `sumeru_addons/...` from your addon’s `init.go` when you need to register hooks or ensure models are loaded.
5. Run `make generate` in the custom workspace after adding or removing addons.
6. Install with `-i your_module` from `sumeru_custom_addons`.

Do **not** copy or patch modules from this repository into a private fork for day-to-day customization.

### Contributors adding a standard business app

Use this repository when the module should ship to every Sumeru deployment:

1. Add a new directory at the repo root named with the technical name.
2. Add `manifest.json` (`depends` only on `base` / other modules here / kernel apps; this Go module must depend only on `sumeru`).
3. Define models as structs with `sdk.Model` embed and `sumeru` tags; run `make generate`.
4. Ship XML under `views/`, ACLs under `security/`, seed data under `data/` as needed.
5. Verify from `sumeru_custom_addons` with `make generate` and `-i <name> --stop-after-init`.

See [CONTRIBUTING.md](CONTRIBUTING.md) for PR expectations.

## Documentation

| Resource | Contents |
| -------- | -------- |
| This README | Role, placement, modules, layout, extend vs contribute |
| [`sumeru/README.md`](https://github.com/ProjectMeru/sumeru/blob/main/README.md) | Core engine, config, CLI |
| [`sumeru_custom_addons/README.md`](https://github.com/ProjectMeru/sumeru_custom_addons/blob/main/README.md) | Workspace runner, `make generate`, custom addons |

## Contributing

See **[CONTRIBUTING.md](CONTRIBUTING.md)** for where to put changes, the generate/test loop, and PR expectations.

Please follow the **[Code of Conduct](CODE_OF_CONDUCT.md)**.

## Security

Report vulnerabilities privately. See **[SECURITY.md](SECURITY.md)**. Do not open public issues for undisclosed security problems.

## License

Licensed under the [Apache License, Version 2.0](LICENSE).
