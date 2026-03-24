# Stack Detection Reference

Use these tables when `blueprint detect-stack` is unavailable or you need to verify its output manually.

## Language Detection (root files)

| File | Language |
|------|----------|
| `composer.json` | PHP |
| `package.json` | Node / JavaScript / TypeScript |
| `go.mod` | Go |
| `Gemfile` | Ruby |
| `requirements.txt` / `pyproject.toml` / `setup.py` | Python |
| `Cargo.toml` | Rust |
| `pom.xml` / `build.gradle` | Java / Kotlin |
| `*.csproj` / `*.sln` | C# / .NET |
| `mix.exs` | Elixir |
| `pubspec.yaml` | Dart / Flutter |

If multiple are found, the project is polyglot — note all languages detected.

## Framework + Version Detection (lock files)

**PHP** (`composer.lock`):
- `laravel/framework` → Laravel (check version)
- `livewire/livewire` → Livewire
- `filament/filament` → Filament
- `symfony/symfony` → Symfony
- `laravel/jetstream`, `laravel/breeze` → auth scaffolding

**Node/JS** (`package-lock.json` / `yarn.lock` / `pnpm-lock.yaml`):
- `next` → Next.js
- `react` → React
- `vue` → Vue
- `@angular/core` → Angular
- `svelte` → Svelte / SvelteKit
- `express` → Express
- `nuxt` → Nuxt
- `astro` → Astro
- `expo` → Expo / React Native

**Go** (`go.sum`):
- `github.com/gin-gonic/gin` → Gin
- `github.com/labstack/echo` → Echo
- `github.com/gofiber/fiber` → Fiber

**Ruby** (`Gemfile.lock`):
- `rails` → Ruby on Rails

**Python** (`requirements.txt` / `pyproject.toml`):
- `django` → Django
- `flask` → Flask
- `fastapi` → FastAPI

**Rust** (`Cargo.toml`):
- `actix-web` → Actix
- `axum` → Axum
- `rocket` → Rocket

## Test Runner Detection

| Marker | Test Runner |
|--------|-------------|
| `pest` in composer.json (require-dev) | Pest PHP |
| `phpunit/phpunit` in composer.json | PHPUnit |
| `jest` in package.json | Jest |
| `vitest` in package.json | Vitest |
| `cypress` in package.json | Cypress |
| `playwright` in package.json | Playwright |
| `pytest` in requirements.txt / pyproject.toml | pytest |
| `_test.go` files exist | Go test |
| `rspec` in Gemfile | RSpec |
| `minitest` in Gemfile | Minitest |
| `#[cfg(test)]` in Rust files | Cargo test |

## Asset Pipeline Detection

| Marker | Tool |
|--------|------|
| `vite.config.*` | Vite |
| `webpack.config.*` | Webpack |
| `esbuild` in package.json | esbuild |
| `tailwind.config.*` | Tailwind CSS |
| `postcss.config.*` | PostCSS |

## Database Detection

Check `.env`, `.env.example`, or config files for:
- `DB_CONNECTION=mysql` or `DATABASE_URL=mysql` → MySQL
- `DB_CONNECTION=pgsql` or `DATABASE_URL=postgres` → PostgreSQL
- `DB_CONNECTION=sqlite` → SQLite
- `MONGODB_URI` or `mongoose` in package.json → MongoDB
- `REDIS_HOST` or `redis` dependency → Redis

## Framework-Specific Directory Templates

Only generate CLAUDE.md files for directories that **actually exist** in the project.

### Laravel

| Directory | Key Content |
|-----------|-------------|
| `app/Models/` | Relationships, factories, casts, scopes, never raw SQL |
| `app/Http/Controllers/` | Single responsibility, use Form Requests, resource controllers |
| `app/Http/Requests/` | Validation rules, authorize method, custom messages |
| `app/Services/` | Business logic lives here, not in controllers |
| `app/Livewire/` | Component patterns, wire:model, events, lifecycle |
| `database/migrations/` | Never migrate:fresh, always add new migrations, never modify existing |
| `database/factories/` | Factory patterns, states, relationships |
| `tests/` | Pest patterns, factories over fixtures, no mocking DB |
| `resources/views/` | Blade/Livewire patterns, component library (DaisyUI if detected) |
| `routes/` | Route naming, middleware, group patterns |
| `config/` | Never hardcode — use env(), config caching |

### Next.js

| Directory | Key Content |
|-----------|-------------|
| `app/` | App Router conventions, server vs client components, layouts, loading/error |
| `components/` | Component naming, props patterns, composition |
| `lib/` | Utility functions, API clients, shared logic |
| `public/` | Static assets only, no sensitive files |
| `tests/` or `__tests__/` | Testing framework, component testing patterns |
| `styles/` | CSS modules or Tailwind patterns |
| `types/` | TypeScript interfaces, shared types |

### Django

| Directory | Key Content |
|-----------|-------------|
| `apps/` or app directories | App structure, models, views, serializers |
| `templates/` | Template inheritance, block patterns |
| `static/` | Static file handling, collectstatic |
| `tests/` | pytest fixtures, factory_boy, API test patterns |
| `core/` or `config/` | Settings structure, URL configuration |

### Go

| Directory | Key Content |
|-----------|-------------|
| `cmd/` | Entry points, CLI structure, flag parsing |
| `internal/` | Package patterns, interfaces, dependency injection |
| `pkg/` | Public packages, API stability |
| `api/` | API definitions, proto files, OpenAPI specs |
| `tests/` or `*_test.go` | Table-driven tests, test helpers, testify patterns |

### React Native / Expo

| Directory | Key Content |
|-----------|-------------|
| `app/` | Expo Router, screen patterns, layouts |
| `components/` | Component patterns, StyleSheet conventions |
| `hooks/` | Custom hooks, state management |
| `services/` or `lib/` | API clients, storage, utilities |
| `__tests__/` | Jest + React Native Testing Library patterns |

### Rails

| Directory | Key Content |
|-----------|-------------|
| `app/models/` | ActiveRecord patterns, validations, scopes |
| `app/controllers/` | Strong params, before_actions, REST conventions |
| `app/views/` | Partials, helpers, Turbo/Hotwire if detected |
| `spec/` or `test/` | RSpec/Minitest patterns, FactoryBot |
| `db/migrations/` | Migration safety, never modify existing |

### Generic (Unknown Framework)

Generate CLAUDE.md only for directories with:
- Complex patterns (5+ related files)
- Test directories
- Configuration directories
- API / integration directories
