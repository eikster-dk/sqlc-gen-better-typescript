# Plan: Native TypeScript Builder

> Source PRD: Conversation refinement -- native TypeScript builder with Zod validation and pg driver

## Architectural decisions

Durable decisions that apply across all phases:

- **Builder name**: `"native"`, the default when `builder` is omitted from config
- **Config shape**: Flat struct with new fields `driver` (default `"pg"`) and `validator` (default `"zod"`). `import_extension` defaults to `".js"`. Irrelevant fields per builder are ignored, not validated.
- **Generated file structure**: `{name}Queries.ts`, `{name}Requests.ts`, `{name}Responses.ts`, plus shared `models.ts`. File naming uses SQL filename as-is (no singularization), converted to camelCase.
- **Shared types in `models.ts`**: `SqlClient` interface (satisfied by `Pool`, `Client`, `PoolClient` from `pg`), `QueryResult<T>` discriminated union, enum literal unions, `ZodError` import.
- **Function signatures**: Plain exported async functions -- `async function queryName(client: SqlClient, params: Params): Promise<QueryResult<T>>`
- **Validation**: Zod `safeParse` on both input (params) and output (results). Errors are discriminated by `phase: "input" | "output"`. Input validation fails fast before any DB IO.
- **Return type**: `QueryResult<T> = { success: true; data: T } | { success: false; error: ZodError; phase: "input" | "output" }`
- **Query command mapping**: `:one` -> `QueryResult<T | null>`, `:many` -> `QueryResult<T[]>`, `:exec` -> `QueryResult<void>`, `:execrows` -> `QueryResult<number>`
- **SQL execution**: Parameterized queries only (`$1, $2, ...`). No template literals.
- **PG-to-Zod type mapping**: `z.coerce.bigint()` for bigint, `z.string().uuid()` for uuid, `z.union([z.literal(...)])` for enums, `z.array()` for array types, `.optional()` for nullable params, `.nullable()` for nullable results.
- **Deferred**: `sqlc.embed()` support, `sqlc.in()` macro expansion.
- **Example project**: `examples/native/`, same schema and queries as effect-v4 (minus embed), vitest + testcontainers + pg Pool.

---

## Phase 1: Config + builder scaffolding

**User stories**: Plugin recognizes the native builder, config accepts new fields, factory returns the correct builder.

### What to build

Extend the plugin config to support the native builder as the default. Add `driver` and `validator` fields with defaults. Wire a new native builder package into the factory so that `builders.NewBuilder(cfg)` returns it. The builder itself is a stub that returns no files -- just enough to prove the wiring works end-to-end from config through factory.

### Acceptance criteria

- [x] `builder` defaults to `"native"` when omitted from config
- [x] `driver` defaults to `"pg"` when omitted
- [x] `validator` defaults to `"zod"` when omitted
- [x] `import_extension` defaults to `".js"` when omitted
- [x] `"native"` is a valid builder in `ValidBuilders`
- [x] `builders.NewBuilder(cfg)` returns a native builder instance when `builder` is `"native"`
- [x] Native builder implements the `Builder` interface and returns an empty file list without error
- [x] Existing `effect-v4-unstable` builder continues to work unchanged
- [x] Go tests pass

---

## Phase 2: models.ts generation

**User stories**: The native builder generates a shared `models.ts` containing foundational types that all query files will depend on.

### What to build

Implement `models.ts` generation in the native builder. This file contains the `SqlClient` interface (a minimal query interface satisfied by `pg`'s `Pool`, `Client`, and `PoolClient`), the `QueryResult<T>` discriminated union type with `phase` discriminator, and the `ZodError` import from Zod. No query generation yet -- just the shared foundation that all subsequent files will import from.

### Acceptance criteria

- [x] Native builder generates a `models.ts` file
- [x] `models.ts` exports a `SqlClient` interface with a `query` method matching `pg`'s signature
- [x] `models.ts` exports a `QueryResult<T>` type as a discriminated union: `{ success: true; data: T } | { success: false; error: ZodError; phase: "input" | "output" }`
- [x] `models.ts` imports `ZodError` from `"zod"`
- [x] Import extensions in `models.ts` respect the `import_extension` config
- [x] Go tests pass

---

## Phase 3: Tracer bullet -- single `:one` query end-to-end

**User stories**: A single `:one` query generates fully working TypeScript with Zod validation, and passes an integration test against a real PostgreSQL database.

### What to build

Implement the core generation pipeline for a single `:one` query. This means: Zod schema generation for basic types (string, number, boolean), request file template (param Zod schemas + inferred types), response file template (result Zod schemas + inferred types), queries file template (async function that safeParse-validates params, executes a parameterized query, safeParse-validates the result, returns `QueryResult`). Create the `examples/native/` project with one table, one `:one` query, testcontainers setup, and one passing integration test proving the full loop works.

### Acceptance criteria

- [x] Native builder generates `{name}Requests.ts` with Zod schemas and inferred types for query params
- [x] Native builder generates `{name}Responses.ts` with Zod schemas and inferred types for query results
- [x] Native builder generates `{name}Queries.ts` with an async function for the `:one` query
- [x] Generated function accepts `SqlClient` and typed params, returns `Promise<QueryResult<T | null>>`
- [x] Input params are validated with `safeParse` before query execution
- [x] Output results are validated with `safeParse` after query execution
- [x] Validation errors include `phase: "input"` or `phase: "output"` accordingly
- [x] Successful `:one` returns `{ success: true, data: T }` when row found
- [x] Successful `:one` returns `{ success: true, data: null }` when no row found
- [x] `examples/native/` project exists with vitest, testcontainers, pg, one table, one query
- [x] Integration test passes against real PostgreSQL

---

## Phase 4: All query commands

**User stories**: All four sqlc query commands (`:one`, `:many`, `:exec`, `:execrows`) are supported with correct return types and validation.

### What to build

Extend the native builder templates and query view construction to support `:many`, `:exec`, and `:execrows` commands alongside the existing `:one`. Each command maps to its agreed return type. Add queries and integration tests to the example project covering each command type.

### Acceptance criteria

- [x] `:many` generates a function returning `Promise<QueryResult<T[]>>`
- [x] `:exec` generates a function returning `Promise<QueryResult<void>>`
- [x] `:execrows` generates a function returning `Promise<QueryResult<number>>`
- [x] All four command types validate input params with `safeParse`
- [x] `:many` and `:one` validate output with `safeParse`
- [x] Integration tests cover all four command types
- [x] All tests pass

---

## Phase 5: Complete PG-to-Zod type mapping

**User stories**: All PostgreSQL column types are correctly mapped to Zod schemas, including edge cases like bigint, uuid, enums, arrays, and nullable columns.

### What to build

Implement the full type mapping from PostgreSQL types to Zod schemas. This includes: `z.coerce.bigint()` for bigint/int8/bigserial, `z.string().uuid()` for uuid, `z.union([z.literal("a"), z.literal("b")])` for user-defined enums, `z.array(...)` for array types, `z.date()` for date/timestamp/timestamptz, `z.unknown()` for json/jsonb, `z.instanceof(Buffer)` for bytea, `z.string()` for numeric/money. Nullable params use `.optional()`, nullable results use `.nullable()`. Add queries and tests exercising each type mapping.

### Acceptance criteria

- [ ] `integer`, `serial`, `int4`, `smallint` map to `z.number()`
- [ ] `bigint`, `int8`, `bigserial` map to `z.coerce.bigint()`
- [ ] `float`, `double precision`, `real` map to `z.number()`
- [ ] `text`, `varchar`, `char` map to `z.string()`
- [ ] `uuid` maps to `z.string().uuid()`
- [ ] `boolean` maps to `z.boolean()`
- [ ] `json`, `jsonb` map to `z.unknown()`
- [ ] `bytea` maps to `z.instanceof(Buffer)`
- [ ] `date`, `timestamp`, `timestamptz` map to `z.date()`
- [ ] `numeric`, `money` map to `z.string()`
- [ ] User-defined enums map to `z.union([z.literal(...), ...])`
- [ ] Array types map to `z.array(...)` wrapping the base type
- [ ] Nullable params generate `.optional()`
- [ ] Nullable results generate `.nullable()`
- [ ] Integration tests cover type edge cases
- [ ] All tests pass

---

## Phase 6: Full example parity with effect-v4

**User stories**: The native builder example project matches the effect-v4 example in schema, queries, and test coverage, proving the builder handles real-world complexity.

### What to build

Expand `examples/native/` to use the same database schema and queries as `examples/effect-v4/` (excluding embed-related queries). Copy over the full seed data with deterministic timestamps. Build a comprehensive integration test suite covering: CRUD operations, pagination, search, count, enum filtering, date ranges, order lines, product sales stats, transactions (commit, rollback, constraint violations), array params, and Zod validation error paths (both input and output phase).

### Acceptance criteria

- [ ] Same database schema as effect-v4 example (customers, products, orders, order_lines, enums)
- [ ] Same queries as effect-v4 example minus embed queries
- [ ] Deterministic seed data matching effect-v4
- [ ] Tests cover CRUD operations
- [ ] Tests cover pagination and search
- [ ] Tests cover enum filtering and date ranges
- [ ] Tests cover transactions (commit, rollback, constraint violations)
- [ ] Tests cover array params
- [ ] Tests cover Zod validation error paths for both input and output phases
- [ ] All tests pass
