# Podux Agent Development Guide

## Project Scope and Technology Stack

Podux is a web management platform for centrally managing frpc server configurations, proxies, connection status, network latency, and runtime logs. The current codebase uses Go 1.25.5, PocketBase 0.35.0, and embedded frp 0.68.0. The application frontend is under `site/` and uses React 19, TypeScript 5.9, Vite 7, Radix Themes 3, and Tailwind CSS 4. `docs/` is a VitePress 1 documentation site; its Vue dependency does not mean that the application frontend uses Vue.

Treat `go.mod`, `site/package.json`, and `docs/package.json` as the sources of truth for versions and dependencies. Do not infer versions from documentation examples.

## Directory Map

| Path | Responsibility |
| --- | --- |
| `main.go` | PocketBase startup and dependency wiring, hooks/custom routes, monitoring scheduler, and static asset serving entry point |
| `internal/domain/` | Server/proxy domain models and repository interfaces |
| `internal/application/` | Use-case services for dashboard, frpc, importer, monitoring, proxy, server, system, version, and related features |
| `internal/infrastructure/persistence/` | Repository implementations based on PocketBase/dbx |
| `internal/interfaces/http/` | Custom HTTP handlers and authentication middleware |
| `migrations/` | The sole version-controlled source of truth for PocketBase collection schemas |
| `pkg/` | Build information, responses, shared types, and utilities |
| `site/` | React application frontend; calls both the PocketBase collection API and custom `/api/*` endpoints |
| `docs/` | Independent VitePress documentation site |
| `development/` | Development design documents for agents and maintainers; not part of the web documentation site |
| `pb_public/` | Copy destination for `site/dist`; embedded into the Go binary by `main.go` via `go:embed`; not committed |
| `pb_data/` | PocketBase SQLite data, uploads, and `frpc/<server-id>` runtime logs/certificates; not committed |
| `build/`, `deploy/` | Multi-platform build scripts and Docker deployment files |

See the [development design documentation](development/README.md) for the full design index.

## Local Development and Verification

Prerequisites: Go 1.25.5 as required by `go.mod`, Node.js 20 (CI baseline), and pnpm 10 (CI baseline).

```bash
# Backend (ensure pb_public/index.html exists first; see the full frontend build below)
go run . serve

# Application frontend; Vite proxies /api to 127.0.0.1:8090
pnpm --dir site install --frozen-lockfile
pnpm --dir site run dev

# Documentation site
pnpm --dir docs install --frozen-lockfile
pnpm --dir docs run docs:dev
```

Run before submitting:

```bash
pnpm --dir site install --frozen-lockfile
pnpm --dir site run lint
pnpm --dir site run build
pnpm --dir docs install --frozen-lockfile
pnpm --dir docs run docs:build

# go:embed must be satisfied before Go tooling loads the main package
rm -rf pb_public
mkdir -p pb_public
cp -R site/dist/. pb_public/
go vet ./...
go test ./...
go build ./...
```

`build/build.sh` builds the frontend, copies `site/dist` to `pb_public`, runs `go mod tidy`, and builds platform packages. It rewrites generated directories and may modify module files, so it is not the preferred verification command for documentation-only or narrowly scoped changes.

## Change Conventions

- Backend: preserve the `interfaces -> application -> domain` dependency direction. Infrastructure implements domain repositories, and `main.go` performs central wiring. Handlers own transport-level validation and responses; business orchestration belongs in application services.
- Frontend: keep React + TypeScript and follow the existing split between `index.tsx` (state/orchestration) and `*.view.tsx` (presentation). Prefer components from `site/src/components/` and Radix Themes. Reuse `site/src/lib/api.ts` or `pocketbase.ts` for API requests.
- Migrations: schema changes must add a reviewable, reversible `migrations/*.go` migration. Migrations are the source of truth; never replace one by modifying a live SQLite database or production schema through the admin UI.
- API: place custom routes in `internal/interfaces/http/` and use `requireAuth` by default. Review PocketBase collection rules as well. When paths, request/response contracts, authentication, or collection fields change, update the relevant design documents.
- Styling: follow the [UI specification](development/ui-spec.md). Prefer Radix tokens, component props, and Tailwind utilities; do not introduce another unsupported set of brand tokens.
- Internationalization: update both `site/src/locales/zh.json` and `en.json` for user-visible text, and access it through `react-i18next`'s `t()`. Avoid new hard-coded strings.
- Documentation: implementation changes must update the [architecture](development/architecture.md), [database design](development/database-design.md), or [UI specification](development/ui-spec.md) when applicable. Documentation claims must link to repository paths or be marked "To be confirmed."

## Security Boundaries

- Do not commit secrets, tokens, certificates, user data, `pb_data/`, `pb_public/`, `site/dist/`, `docs/.vitepress/dist/`, or build packages.
- `fh_servers.auth`, TLS private keys, user passwords, and token keys are sensitive. Never log their original values. Avoid logging, copying, or exposing tokens used in SSE query parameters.
- Do not substitute manual database operations for migrations. Relations currently have no cascading deletes; assess orphaned-data risks before deleting parent records.
- Do not perform unrelated repository-wide formatting or overwrite existing worktree changes.
- Changes to endpoints, authentication rules, or data structures must update tests, callers, and the design documents in this directory.

## Pre-Submission Checklist

- [ ] `git diff` contains only the intended scope, with no runtime data, build artifacts, or unrelated formatting.
- [ ] Go vet/test/build, frontend lint/build, and documentation build were run and their results reported accurately.
- [ ] New or changed APIs, collections, relations, and rules are reflected in the database/architecture documentation.
- [ ] The UI reuses shared components; both themes, the 760px mobile layout, Chinese/English text, and basic keyboard operation were checked.
- [ ] Relative Markdown links work, and Mermaid names match the code and render correctly.
- [ ] No sensitive values appear in code, logs, screenshots, or commit history.
