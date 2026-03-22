# sqlc-gen-better-typescript

A [sqlc](https://sqlc.dev) WASM plugin that generates type-safe TypeScript code from your SQL queries.

## Requirements

- [sqlc](https://sqlc.dev) v1.25.0 or later
- For the `effect-v4-unstable` builder:
  - [Effect](https://effect.website) v4 (beta)
  - TypeScript 5.5+

## What is this?

**sqlc-gen-better-typescript** is a flexible TypeScript code generator for sqlc that supports multiple output formats through a builder architecture. Instead of writing boilerplate database access code, you write SQL and the plugin generates fully typed TypeScript code tailored to your preferred libraries and patterns.

The current focus is on [Effect v4](https://effect.website) code generation, with planned support for:
- Native TypeScript (no external dependencies)
- Zod v4 schema validation
- Effect v3 compatibility

Depending on the builder you choose, you get:

- **Type-safe parameter schemas** using Effect's Schema library
- **Type-safe result schemas** with proper null handling via `Option`
- **Repository services** with Effect's dependency injection via `Layer`
- **Automatic SQL type mapping** to TypeScript/Effect types

### Effect v4 example

Write a SQL query:

```sql
-- name: GetCustomer :one
SELECT id, email, name, phone, created_at
FROM customers
WHERE id = $1;
```

Get a fully typed Effect repository:

```typescript
// Generated automatically
export const GetCustomerParams = Schema.Struct({
  id: Schema.Int,
})

export type GetCustomerParams = typeof GetCustomerParams.Type

export const GetCustomerResult = Schema.Struct({
  id: Schema.Int,
  email: Schema.String,
  name: Schema.String,
  phone: Schema.OptionFromNullOr(Schema.String),
  created_at: Schema.Date,
})

export type GetCustomerResult = typeof GetCustomerResult.Type

// Repository interface
export interface CustomersRepositoryShape {
  readonly getCustomer: (params: GetCustomerParams) => Effect.Effect<
    Option.Option<GetCustomerResult>,
    SqlError.SqlError | Schema.SchemaError
  >
}

// Service Tag
export class CustomersRepository extends ServiceMap.Service<
  CustomersRepository,
  CustomersRepositoryShape
>()("CustomersRepository") {}

// Implementation
const customersRepositoryImpl = Effect.gen(function* () {
  const sql = yield* SqlClient.SqlClient

  const getCustomer = SqlSchema.findOneOption({
    Request: GetCustomerParams,
    Result: GetCustomerResult,
    execute: (params) => sql.unsafe(
      `SELECT id, email, name, phone, created_at FROM customers WHERE id = $1`,
      [params.id]
    )
  })

  return { getCustomer } satisfies CustomersRepositoryShape
})

// Live Layer
export const customersRepositoryLive = Layer.effect(CustomersRepository, customersRepositoryImpl)

// Usage
const program = Effect.gen(function* () {
  const repo = yield* CustomersRepository
  const customer = yield* repo.getCustomer({ id: 1 })
  // customer is Option.Option<GetCustomerResult>
})
```

## Builders

The plugin uses a **builder** architecture to support different code generation targets. Each builder produces output tailored for a specific framework or library version.

## Available Builders

| Builder | Description | Status |
|---------|-------------|--------|
| `effect-v4-unstable` | Generates Effect v4 TypeScript code using `effect/unstable/sql` | Available |

### Effect v4 Builder

The `effect-v4-unstable` builder generates idiomatic Effect v4 code using the `effect/unstable/sql` module.

#### SQL Injection Safety

The generated code uses `sql.unsafe()` from Effect's SQL module, but **this is safe from SQL injection**. Despite the name, `sql.unsafe()` simply indicates that the query string is not built using Effect's tagged template literal. The generated queries use PostgreSQL's parameterized placeholders (`$1`, `$2`, etc.) with values passed as a separate array:

```typescript
sql.unsafe(
  `SELECT * FROM customers WHERE id = $1 AND email = $2`,
  [params.id, params.email]  // Values are safely parameterized
)
```

The parameters are never interpolated into the SQL string - they are sent separately to PostgreSQL, which handles them safely. This is the same protection you get from prepared statements.

#### Template Literals

By default, the plugin generates code using Effect's `sql` tagged template literal:

```typescript
// Default output (template literals)
// GetCustomer
// SELECT * FROM customers WHERE id = $1 AND email = $2
execute: (params) => sql`SELECT * FROM customers WHERE id = ${params.id} AND email = ${params.email}`
```

The original SQL query is included as a comment above each query implementation for reference.

If you need to use `sql.unsafe()` instead (e.g., for compatibility reasons), you can disable template literals:

```yaml
options:
  builder: effect-v4-unstable
  disable_template_literals: true
```

This generates:

```typescript
// With disable_template_literals: true
// GetCustomer
// SELECT * FROM customers WHERE id = $1 AND email = $2
execute: (params) => sql.unsafe(
  `SELECT * FROM customers WHERE id = $1 AND email = $2`,
  [params.id, params.email]
)
```

Both approaches are equally safe from SQL injection - this is purely a stylistic preference.

#### Repository Pattern

Each SQL file in your `queries/` directory becomes its own encapsulated repository. For example:

```
queries/
├── customers.sql    → CustomersRepository.ts
├── orders.sql       → OrdersRepository.ts
└── products.sql     → ProductsRepository.ts
```

All queries defined in a SQL file are grouped into a single repository service. This keeps related database operations together and provides clean dependency injection through Effect's `Layer` system.

#### Generated Output

For each repository, the builder generates:

- **Parameter schemas** - Type-safe input validation for each query
- **Result schemas** - Type-safe output parsing with proper null handling (`Option`)
- **Repository interface** - A typed interface defining all available operations
- **Service tag** - An Effect service tag for dependency injection
- **Live implementation** - The actual repository implementation using `SqlClient`
- **Layer export** - A ready-to-use `Layer` for providing the repository

#### Usage

```typescript
import { CustomersRepository, customersRepositoryLive } from "./repositories/CustomersRepository"
import { Effect, Layer } from "effect"
import { PgClient } from "effect/unstable/sql/PgClient"

const program = Effect.gen(function* () {
  const repo = yield* CustomersRepository
  
  // All queries from customers.sql are available as methods
  const customer = yield* repo.getCustomer({ id: 1 })
  const allCustomers = yield* repo.listCustomers()
  yield* repo.createCustomer({ email: "new@example.com", name: "New Customer" })
})

// Provide the repository layer (requires SqlClient)
const runnable = program.pipe(
  Effect.provide(customersRepositoryLive),
  Effect.provide(/* your PgClient layer */)
)
```

#### Supported sqlc Commands

| Command | Supported | Effect Return Type | Description |
|---------|-----------|-------------------|-------------|
| `:one` | Yes | `Option.Option<Result>` | Returns at most one row |
| `:many` | Yes | `Result[]` | Returns zero or more rows |
| `:exec` | Yes | `void` | Executes without returning data |
| `:execrows` | Yes | `number` | Returns the number of affected rows |
| `:execresult` | No | - | Not yet implemented |
| `:copyfrom` | No | - | Not yet implemented |
| `:batchexec` | No | - | Not yet implemented |
| `:batchone` | No | - | Not yet implemented |
| `:batchmany` | No | - | Not yet implemented |

#### Supported sqlc Macros

| Macro | Supported | Description |
|-------|-----------|-------------|
| `sqlc.arg('name')` | Yes | Explicit parameter naming |
| `sqlc.narg('name')` | No | Nullable argument - not yet implemented |
| `sqlc.slice('name')` | No | Slice expansion - use `= ANY($1::type[])` instead (see below) |
| `sqlc.embed(table)` | No | Embed table columns - not yet implemented |

> **Note on `sqlc.slice`:** While `sqlc.slice()` is not supported, you can achieve the same result using PostgreSQL's `ANY` operator with array casting:
> ```sql
> -- Instead of: WHERE id IN (sqlc.slice('ids'))
> -- Use:
> WHERE id = ANY($1::int[])
> ```
> This generates a parameter typed as `Schema.Array(Schema.Int)` and works correctly with PostgreSQL.

### Future Builders (Planned)

| Builder | Description |
|---------|-------------|
| `effect-v3` | Effect v3 compatible code generation |
| `typescript` | Plain TypeScript with no Effect dependency |
| `zod-v4` | TypeScript with Zod v4 schemas for validation |

## Supported Database Engines

| Engine | Supported |
|--------|-----------|
| PostgreSQL | Yes |
| MySQL | No |
| SQLite | No |

## Type Mapping

The following table shows how PostgreSQL types are mapped to Effect Schema types:

| PostgreSQL Type | Effect Schema | Notes |
|-----------------|---------------|-------|
| `integer`, `int`, `int4`, `serial` | `Schema.Int` | |
| `bigint`, `int8`, `bigserial` | `BigIntFromString` | PostgreSQL returns bigint as string to preserve precision |
| `smallint`, `int2`, `smallserial` | `Schema.Int` | |
| `real`, `float4`, `double precision`, `float8` | `Schema.Number` | |
| `numeric`, `money` | `Schema.String` | Preserves precision |
| `boolean`, `bool` | `Schema.Boolean` | |
| `text`, `varchar`, `char`, `citext` | `Schema.String` | |
| `uuid` | `Schema.String` | |
| `date` | `Schema.Date` | |
| `timestamp`, `timestamptz` | `Schema.Date` | |
| `time`, `timetz`, `interval` | `Schema.String` | |
| `json`, `jsonb` | `Schema.Unknown` | |
| `bytea` | `Schema.Uint8Array` | |
| `inet`, `cidr`, `macaddr` | `Schema.String` | |
| Arrays (e.g., `int[]`) | `Schema.Array(...)` | Wraps the base type |
| Enums | `Schema.Literals([...])` | Generated from enum definition |

### Nullability

- **Parameters**: Nullable parameters use `Schema.optional()`, allowing callers to omit the field
- **Results**: Nullable results use `Schema.OptionFromNullOr()`, transforming `null` to `Option.None`

## Configuration

Configure the plugin in your `sqlc.yaml`:

```yaml
version: '2'
plugins:
- name: better-typescript
  wasm:
    url: https://github.com/eikster-dk/sqlc-gen-better-typescript/releases/download/v[version]/plugin.wasm
    sha256: [calculatedSha]

sql:
- schema: schema/
  queries: queries/
  engine: postgresql
  codegen:
  - out: src/repositories
    plugin: better-typescript
    options:
      builder: effect-v4-unstable
      # debug: true
      # debug_dir: debug
```

### Plugin Options

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `builder` | string | Yes | - | The code generation builder to use. Must be one of the available builders (e.g., `effect-v4-unstable`). |
| `disable_template_literals` | boolean | No | `false` | Disable template literals and use `sql.unsafe()` instead. See [Template Literals](#template-literals). |
| `debug` | boolean | No | `false` | Enable debug mode to output intermediate representations and detailed logs during code generation. |
| `debug_dir` | string | No | `"debug"` | Directory where debug output files are written when debug mode is enabled. |

## Getting Started

1. Install sqlc: https://docs.sqlc.dev/en/latest/overview/install.html

2. Create your `sqlc.yaml` configuration (see above)

3. Write your SQL schema and queries

4. Run sqlc:
   ```bash
   sqlc generate
   ```

5. Use the generated repositories in your Effect application

## Development

### Building the Plugin

```bash
make build
```

### Running Tests

```bash
make test
```

### Project Structure

```
.
├── cmd/plugin/           # Plugin source code
│   ├── main.go           # Entry point
│   └── internal/
│       ├── builders/     # Code generation builders
│       ├── config/       # Plugin configuration
│       ├── mapper/       # sqlc to internal type mapping
│       ├── models/       # Internal data models
│       └── logger/       # Structured logging
├── examples/             # Example projects
│   └── effect-v4/        # Effect v4 example
└── dist/                 # Built plugin artifacts
```

## License

See [LICENSE](LICENSE) file.
