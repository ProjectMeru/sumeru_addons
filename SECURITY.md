# Security Policy

## Supported versions

Sumeru addons are experimental. Security fixes are applied to the latest **`main`** branch of [ProjectMeru/sumeru_addons](https://github.com/ProjectMeru/sumeru_addons). Older commits or forks are not regularly backported unless maintainers announce otherwise.

Related repositories (`sumeru`, `sumeru_custom_addons`) should follow the same reporting process when the issue spans those trees; report against the repo that owns the vulnerable code when possible.

## Reporting a vulnerability

**Please do not open a public GitHub issue** for undisclosed security problems.

Prefer **GitHub Security Advisories** (private vulnerability reporting) on this repository:

https://github.com/ProjectMeru/sumeru_addons/security/advisories/new

If private reporting is unavailable, contact the **ProjectMeru** organization maintainers via GitHub (https://github.com/ProjectMeru) and mark the communication as security-sensitive.

Include as much of the following as you can:

- Description of the issue and impact
- Steps to reproduce or a minimal proof of concept
- Affected version / commit / branch
- Any suggested fix (optional)

We will acknowledge valid reports as promptly as practical, assess impact, and coordinate disclosure after a fix is available when appropriate.

## Scope

**In scope (examples):**

- Access control / RBAC bypasses in addon security XML or `sys.access.csv`
- Unsafe handling of record data or events in addon Go code (e.g. privilege bypass via `orm` helpers)
- Injection or XSS introduced by addon views, templates, or static assets shipped here
- Known vulnerable dependencies that are reachable through these addons in default or documented configurations

**Out of scope (examples):**

- Misconfiguration (e.g. weak `db_password` or `db_sslmode = disable` in local/example configs)
- Issues that require already-compromised database credentials or host access
- Social engineering, physical attacks, or denial-of-service without a clear application bug
- Vulnerabilities only in third-party forks or unpublished custom addons outside ProjectMeru repositories
- Core engine issues that belong in [ProjectMeru/sumeru](https://github.com/ProjectMeru/sumeru) (report there instead)

## Safe harbor

We appreciate good-faith research. Avoid privacy violations, data destruction, and disruption of production systems you do not own. Test against local or authorized environments only.
