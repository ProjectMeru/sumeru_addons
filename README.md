# sumeru_addons

[![CI](https://github.com/ProjectMeru/sumeru_addons/actions/workflows/ci.yml/badge.svg)](https://github.com/ProjectMeru/sumeru_addons/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.26.6+-00ADD8?logo=go&logoColor=white)](https://go.dev/dl/)
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

**Standard business addons** for [Sumeru](https://github.com/ProjectMeru/sumeru) — CRM, Sales, Accounting, Purchase, HR, and supporting apps. Each directory at the repo root is an installable Sumeru module with a `manifest.json`, Go models, and XML views.

Go module: `sumeru_addons` · depends on [`sumeru`](../sumeru/) only (`replace sumeru => ../sumeru` in `go.mod`).

---

## Module catalog

| Module | Display name | Summary | Depends |
| ------ | -------------- | ------- | ------- |
| [`product`](product/) | Products | Product catalog shared by sales, purchase, and accounting | `base` |
| [`utm`](utm/) | UTM Tracking | Campaigns, mediums, and sources for lead attribution | `base` |
| [`crm`](crm/) | CRM | Leads, pipelines, teams, and opportunities | `base`, `contacts`, `mail`, `utm` |
| [`sale`](sale/) | Sales | Quotations and sales orders | `product`, `contacts`, `crm` |
| [`sale_crm`](sale_crm/) | Sales CRM Bridge | Draft quotation when an opportunity is won | `sale`, `crm` |
| [`account`](account/) | Invoicing | COA, journal entries, taxes, payments, financial reports, bank reconciliation, analytic | `product`, `contacts` |
| [`purchase`](purchase/) | Purchase | RFQ → confirmed PO → vendor bill | `account`, `product` |
| [`hr`](hr/) | Employees | Employee directory, departments, and job positions | `base`, `contacts` |

Kernel apps such as `contacts` and `mail` ship with **[`sumeru`](../sumeru/)** — they are not in this repository but are required by several modules above.

### Install order

Install modules in dependency order. A typical full stack:

```text
product → utm → crm → sale → sale_crm → account → purchase → hr
```

From your Sumeru workspace (see [sumeru_custom_addons](https://github.com/ProjectMeru/sumeru_custom_addons)):

```bash
go run . -- -c sumeru.conf -i product,utm,crm,sale,sale_crm,account,purchase,hr --stop-after-init
```

Update a single module after pulling changes:

```bash
go run . -- -c sumeru.conf -u account --stop-after-init
```

---

## What each app covers

### CRM & Sales

- **CRM** — Pipeline kanban, forecast views, lead scoring (PLS), assignment rules, merge/convert/lost wizards, prorated and recurring revenue fields.
- **Sales** — Quotation → confirmed order, sequences, tax-aware lines, invoice status tracking.
- **sale_crm** — Event bridge: winning an opportunity creates a draft quotation (no UI of its own).

### Accounting

The **account** module is a consolidated accounting suite:

| Area | Models / features |
| ---- | ----------------- |
| Core | Chart of accounts, journals, moves & lines, taxes, payment terms, fiscal positions |
| Payments | Customer/vendor payments, payment register wizard, move reversal |
| Reporting | Profit & Loss, Balance Sheet, Trial Balance, General Ledger |
| Bank | Bank statements, statement lines, reconciliation workspace |
| Analytic | Analytic plans and accounts |

### Operations & HR

- **product** — Products and categories; income/expense account links for posting.
- **purchase** — RFQ/PO workflow with vendor bill generation via `account`.
- **hr** — Employees, departments, and jobs.
- **utm** — Marketing attribution (campaign, medium, source) linked from CRM leads.

---

## Addon anatomy

Every module in this repo follows the same layout:

```text
<technical_name>/
  manifest.json       # name, depends, data files, application flag
  init.go             # blank-imports models, services, controllers, wizards
  models/             # Go structs (sdk.Model embed + sumeru tags)
  views/              # list, form, kanban, actions, menus (XML)
  security/           # groups, record rules, sys.access.csv
  data/               # seed XML, sequences, demo (optional)
  services/           # business logic (optional)
  controllers/        # HTTP routes (optional)
  wizard/             # transient models (optional)
```

Rules:

- **Folder name = technical module name** — must match `^[a-z][a-z0-9_]*$` and equal `manifest.json` → `"name"`.
- **`depends`** — only list modules from `sumeru/addons` or this repo; this Go module must not import other addon repos.
- **`data`** — paths relative to the module root; loaded on `-i` (install) or `-u` (update).

---

## Go models

Models are Go structs embedding `sdk.Model` with struct tags consumed by the ORM and code generator.

```go
type SaleOrder struct {
    sdk.Model `sumeru:"model=sale.order"`

    PartnerID sdk.Many2One[CorePartner]       `sumeru:"string=Customer"`
    State     sdk.Selection[SaleOrderState]    `sumeru:"string=Status,default=draft"`
    AmountTotal sdk.Numeric                    `sumeru:"string=Total,precision=18,scale=2,default=0"`
}
```

Conventions used across this repo:

| Pattern | Location | Purpose |
| ------- | -------- | ------- |
| `selection_types.go` | `models/` | Typed enums for `sdk.Selection[T]` |
| `zrefs.go` | `models/` | Aliases for cross-module relations (`CorePartner`, `ProductProduct`, …) |
| `zmodels.go` | `models/`, `wizard/` | Generated `sdk.MustRegister` — run `make generate` from the workspace after adding models |

Internal relations use the module’s own structs (`Many2One[AccountJournal]`). Cross-module refs go through `zrefs.go` aliases, matching the pattern in [`engagement_cookbook`](../sumeru_custom_addons/addons/engagement_cookbook/models/).

---

## Development

### Run and test

This repo is **not** the process entry point. Clone it as a sibling of [`sumeru`](../sumeru/) and [`sumeru_custom_addons`](../sumeru_custom_addons/), add `../sumeru_addons` to `addons_path`, then run the server from the custom workspace. Full setup: **[sumeru_custom_addons README](https://github.com/ProjectMeru/sumeru_custom_addons/blob/main/README.md)**.

After model or module changes:

```bash
# From sumeru_custom_addons
make generate

# From this repo
go test ./...
go vet ./...
```

### Where to put changes

| Goal | Where |
| ---- | ----- |
| Improve a standard business app | Pull request to **this repo** |
| Customer-specific module or override | [`sumeru_custom_addons/addons/`](../sumeru_custom_addons/addons/) |
| Engine, kernel apps (`base`, `mail`, …) | [`sumeru`](../sumeru/) |

Do not fork or patch modules here for a single deployment — extend from the custom workspace so upstream pulls stay clean.

See **[CONTRIBUTING.md](CONTRIBUTING.md)** for the PR workflow, generate loop, and review expectations.

---

## Related repositories

| Repository | Role |
| ---------- | ---- |
| [ProjectMeru/sumeru](https://github.com/ProjectMeru/sumeru) | Core engine and kernel addons |
| [ProjectMeru/sumeru_addons](https://github.com/ProjectMeru/sumeru_addons) | Standard business apps (this repo) |
| [ProjectMeru/sumeru_custom_addons](https://github.com/ProjectMeru/sumeru_custom_addons) | Workspace runner, custom addons, generated imports |

---

## Contributing · Security · License

- **[CONTRIBUTING.md](CONTRIBUTING.md)** — development setup and pull request guidelines
- **[CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)** — community standards
- **[SECURITY.md](SECURITY.md)** — responsible disclosure (do not open public issues for undisclosed vulnerabilities)

Licensed under the [Apache License, Version 2.0](LICENSE).
