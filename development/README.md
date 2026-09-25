# Podux Development Design Documentation

This directory documents the verifiable design of the current `main` baseline and serves as the entry point for future development and agent changes:

- [System architecture](architecture.md): in-process components, layering, request/startup flows, build and deployment, and CI boundaries.
- [Database design](database-design.md): all application collections, relations, rules, indexes, lifecycles, and implementation discrepancies.
- [UI specification](ui-spec.md): themes, actual tokens, layouts, and component reuse rules for the React application frontend.

## Sources of Truth and Scope

These documents are based on repository code and configuration, not speculation about a target architecture. For the database, `migrations/*.go` is the primary source of truth and is cross-checked against domain models, repositories, application services, HTTP handlers, and callers under `site/src/`. For the UI, the sources are `site/src/index.css`, theme contexts, layouts, shared components, and pages. Build and deployment facts come from `go.mod`, package manifests, `main.go`, CI, `build/build.sh`, and Docker files.

"Verified current state" describes the present implementation. "Current discrepancy / To be confirmed" and "Technical debt" identify gaps that the code cannot resolve or constraints that are not yet unified; they do not authorize schema or behavior changes.

## Update Triggers

- Update the architecture document when component boundaries, startup flow, custom routes, background tasks, persistence, or deployment change.
- Update the database document when migrations, collection rules/indexes/relations, field mappings, or retention/deletion policies change.
- Update the UI document when theme configuration, design tokens, layout breakpoints, shared components, interaction states, internationalization, or accessibility constraints change.
- Update the repository-root `AGENT.md` when any of the changes above also affect agent working conventions.
