# sumeru_addons

Standard **business addons** for Sumeru (separate Go module: `module sumeru_addons`).

- Depends on **[`../sumeru`](../sumeru/)** only (`replace sumeru => ../sumeru` in `go.mod`).
- **Do not run the HTTP server from this tree** — use **[`../sumeru_custom_addons`](../sumeru_custom_addons/)** so `addons_path` can list `../sumeru/addons`, `../sumeru_addons`, and your `./addons` together.

## Layout

Each installable app is a **direct child directory** of this repo (sibling to `go.mod`), with `manifest.json` and optional `init.go` / `models/`:

```text
sumeru_addons/
  go.mod
  <technical_name>/
    manifest.json
    init.go              # optional: blank-import models
    models/ …
    views/ …
```

Technical module names must match **`^[a-z][a-z0-9_]*$`** and equal the **folder name**.

## Commercial flow (v1)

| Module | Role |
|--------|------|
| `contacts` | Address book (`core.partner`) |
| `product` | Product catalog |
| `crm` | Leads / opportunities |
| `sale` | Quotations / sales orders |
| `sale_crm` | Opportunity → draft quotation (on Won) |
| `account` | Invoices, bills, COA, post journal lines |
| `purchase` | RFQ / PO → vendor bill |
| `hr` | Employees, departments, jobs |

Typical install order: `contacts, product, crm, sale, sale_crm, account, purchase, hr`.

## Custom extensions

Put overrides and customer-specific modules under **`sumeru_custom_addons/addons/`**, list **`../sumeru_addons`** on **`addons_path`** before **`./addons`**, and blank-import **`sumeru_addons/...`** from your addon’s `init.go` when you need to register hooks or inherit models.
