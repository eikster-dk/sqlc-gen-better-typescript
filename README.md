# sqlc-gen-effect

A [sqlc](https://sqlc.dev) WASM plugin that generates type-safe TypeScript code from your SQL queries.

## Requirements

- [sqlc](https://sqlc.dev) v1.25.0 or later
- For the Effect v4 plugin:
  - [Effect](https://effect.website) v4 (beta)
  - TypeScript 5.5+

## What is this?

**sqlc-gen-effect** contains two sqlc WASM plugins backed by a shared `toolbelt` package that normalizes sqlc plugin requests into a stable intermediate representation. Instead of writing boilerplate database access code, you write SQL and generate typed TypeScript code for either Effect v4 or plain async functions.

Current builders:
- **Effect v4**: repository services using `effect/unstable/sql` and Effect Schema
- **Native TypeScript**: plain async functions using a small `SqlClient` interface and Zod validation

Planned support:
- Zod v4 schema validation
- Effect v3 compatibility

The Effect plugin generates:

- **Type-safe parameter schemas** using Effect's Schema library
- **Type-safe result schemas** with proper null handling via `Option`
- **Repository services** with Effect's dependency injection via `Layer`
- **Automatic SQL type mapping** to TypeScript/Effect types

The native plugin generates:

- **Plain async query functions** grouped by SQL file
- **Zod request/result schemas** and inferred TypeScript types
- **A small database client interface** compatible with `pg.Pool` / `pg.Client`
- **Shared enum schemas** in `models.ts`
- **Nested `sqlc.embed` result transforms** for joined table structures

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

// Implementation
const customersRepositoryMake = Effect.gen(function* () {
  const sql = yield* SqlClient.SqlClient

  const getCustomer = SqlSchema.findOneOption({
    Request: GetCustomerParams,
    Result: GetCustomerResult,
    execute: (params) => sql.unsafe(
      `SELECT id, email, name, phone, created_at FROM customers WHERE id = $1`,
      [params.id]
    )
  })

  return { getCustomer } as const
})

// Service Tag
export class CustomersRepository extends Context.Service<CustomersRepository>()(
  "CustomersRepository",
  { make: customersRepositoryMake }
) {
  static readonly layer = Layer.effect(this, this.make)
}

export type CustomersRepositoryService = CustomersRepository["Service"]

// Live Layer
export const customersRepositoryLive = CustomersRepository.layer

// Usage
const program = Effect.gen(function* () {
  const repo = yield* CustomersRepository
  const customer = yield* repo.getCustomer({ id: 1 })
  // customer is Option.Option<GetCustomerResult>
})
```

### Nested Results with `sqlc.embed`

Use `sqlc.embed` to group columns from joined tables into nested structures. This is useful for queries that join multiple tables and you want the result to reflect that structure.

Write a SQL query with embeds:

```sql
-- name: GetOrderWithCustomer :one
SELECT sqlc.embed(orders), sqlc.embed(customers)
FROM orders
JOIN customers ON orders.customer_id = customers.id
WHERE orders.id = $1;
```

Get a nested result type:

```typescript
// Generated automatically - Row schema represents flat database result
const GetOrderWithCustomerRow = Schema.Struct({
  orders_id: Schema.Int,
  orders_customer_id: Schema.Int,
  orders_status: OrderStatusSchema,
  orders_total_cents: Schema.Int,
  orders_shipping_address: Schema.NullOr(Schema.String),
  // ... more orders columns
  customers_id: Schema.Int,
  customers_email: Schema.String,
  customers_name: Schema.String,
  // ... more customers columns
})

// Nested embed schemas with proper Option handling
const OrderEmbed = Schema.Struct({
  id: Schema.Int,
  customer_id: Schema.Int,
  status: OrderStatusSchema,  // Enums are preserved
  total_cents: Schema.Int,
  shipping_address: Schema.OptionFromNullOr(Schema.String),  // Nullable -> Option
  // ...
})

const CustomerEmbed = Schema.Struct({
  id: Schema.Int,
  email: Schema.String,
  name: Schema.String,
  // ...
})

// Result schema transforms flat rows to nested structure
export const GetOrderWithCustomerResult = GetOrderWithCustomerRow.pipe(
  Schema.decodeTo(
    Schema.Struct({
      order: OrderEmbed,      // Singularized table name
      customer: CustomerEmbed,
    }),
    SchemaTransformation.transform({
      decode: (row) => ({
        order: {
          id: row.orders_id,
          customer_id: row.orders_customer_id,
          status: row.orders_status,
          shipping_address: row.orders_shipping_address,
          // ...
        },
        customer: {
          id: row.customers_id,
          email: row.customers_email,
          name: row.customers_name,
          // ...
        },
      }),
      encode: () => { throw new Error("Encode not supported for sqlc.embed queries"); },
    })
  )
)

// Result type is nested:
// {
//   order: { id: number, status: "pending" | "shipped" | ..., shipping_address: Option<string>, ... }
//   customer: { id: number, email: string, name: string, ... }
// }
```

**Key features:**
- Table names are singularized for field names (`orders` → `order`, `customers` → `customer`)
- Columns are prefixed with table name to avoid conflicts (`orders_id`, `customers_id`)
- Enum types are preserved in embed schemas
- Nullable fields use `Schema.OptionFromNullOr` for consistent API with non-embed queries
- The transformation is decode-only (embed queries are read-only)

## Effect v4 Plugin

The Effect v4 plugin generates idiomatic Effect v4 code using the `effect/unstable/sql` module.

#### SQL Generation

By default, the plugin transforms sqlc's parameterized SQL into Effect's tagged template literal syntax:

```typescript
// Default output (template literals)
// GetCustomer
// SELECT * FROM customers WHERE id = $1 AND email = $2
execute: (params) => sql`SELECT * FROM customers WHERE id = ${params.id} AND email = ${params.email}`
```

The original SQL query is included as a comment above each query implementation for reference.

#### Preserving Original SQL

If you prefer to keep the sqlc-generated SQL statements unmodified, you can disable template literal transformation:

```yaml
options:
  disable_template_literals: true
```

This uses `sql.unsafe()` which passes the SQL exactly as sqlc generated it:

```typescript
// With disable_template_literals: true
// GetCustomer
// SELECT * FROM customers WHERE id = $1 AND email = $2
execute: (params) => sql.unsafe(
  `SELECT * FROM customers WHERE id = $1 AND email = $2`,
  [params.id, params.email]
)
```

Both approaches are equally safe from SQL injection. The choice is between:
- **Template literals (default)**: Cleaner syntax, but transforms the SQL by replacing `$1`, `$2` placeholders with interpolated parameters
- **sql.unsafe()**: Preserves the original sqlc-generated SQL without modification

> **Note:** In both cases, the plugin may still modify SQL to handle edge cases like duplicate column names (e.g., `id` becomes `id`, `id_2`, `id_3`).

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
| `:execresult` | Yes | `SqlExecResult` | Returns command tag and affected row count |
| `:copyfrom` | No | - | Explicitly rejected during generation |
| `:batchexec` | No | - | Explicitly rejected during generation |
| `:batchone` | No | - | Explicitly rejected during generation |
| `:batchmany` | No | - | Explicitly rejected during generation |

#### Supported sqlc Macros

| Macro | Supported | Description |
|-------|-----------|-------------|
| `sqlc.arg('name')` | Yes | Explicit parameter naming |
| `@name` | Yes | PostgreSQL shorthand for `sqlc.arg(name)` |
| `sqlc.narg('name')` | Yes | Explicit nullable parameter naming |
| `sqlc.slice('name')` | Yes | Dynamic `IN` list expansion using Effect SQL's `sql.in` helper |
| `sqlc.embed(table)` | Yes | Embed table columns into nested structures |

`sqlc.slice` generates array request schemas and expands to Effect SQL's column-aware `sql.in` helper. Empty slices compile to a false predicate through Effect SQL.

### Native TypeScript Plugin

The native plugin generates plain TypeScript functions with Zod validation and no Effect dependency. It targets PostgreSQL clients that implement this interface:

```typescript
export interface SqlClient {
  query(queryText: string, values?: unknown[]): Promise<SqlQueryResult>
}
```

For each SQL file, the native builder generates three files:

- `*Requests.ts`: Zod input schemas
- `*Responses.ts`: Zod row/result schemas
- `*Queries.ts`: async query functions

Example generated query:

```typescript
const result = await getCustomer(pool, { id: 1 })

if (result.success) {
  console.log(result.data)
}
```

Native query functions return `QueryResult<T>`:

```typescript
export type QueryResult<T> =
  | { success: true; data: T }
  | { success: false; error: ZodError; phase: "input" | "output" }
```

#### Native Supported sqlc Commands

| Command | Supported | Native Return Type | Description |
|---------|-----------|--------------------|-------------|
| `:one` | Yes | `QueryResult<Result | null>` | Returns at most one row |
| `:many` | Yes | `QueryResult<Result[]>` | Returns zero or more rows |
| `:exec` | Yes | `QueryResult<void>` | Executes without returning data |
| `:execrows` | Yes | `QueryResult<number>` | Returns the number of affected rows |
| `:execresult` | Yes | `QueryResult<SqlExecResult>` | Returns command tag and affected row count |
| `:copyfrom` | No | - | Explicitly rejected during generation |
| `:batchexec` | No | - | Explicitly rejected during generation |
| `:batchone` | No | - | Explicitly rejected during generation |
| `:batchmany` | No | - | Explicitly rejected during generation |

#### Native Supported sqlc Macros

| Macro | Supported | Description |
|-------|-----------|-------------|
| `sqlc.arg('name')` | Yes | Explicit parameter naming |
| `@name` | Yes | PostgreSQL shorthand for `sqlc.arg(name)` |
| `sqlc.narg('name')` | Yes | Explicit nullable parameter naming |
| `sqlc.slice('name')` | Yes | Dynamic `IN` list expansion; empty slices become `IN (NULL)` |
| `sqlc.embed(table)` | Yes | Embed table columns into nested result objects |

### Future Builders (Planned)

| Builder | Description |
|---------|-------------|
| `effect-v3` | Effect v3 compatible code generation |
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
- name: effect
  wasm:
    url: https://github.com/eikster-dk/sqlc-gen-better-typescript/releases/download/v[version]/sqlc-gen-effect.wasm
    sha256: [calculatedSha]
- name: native
  wasm:
    url: https://github.com/eikster-dk/sqlc-gen-better-typescript/releases/download/v[version]/sqlc-gen-native.wasm
    sha256: [calculatedSha]

sql:
- schema: schema/
  queries: queries/
  engine: postgresql
  codegen:
  - out: src/repositories
    plugin: effect
    options:
      # debug: true
      # debug_dir: debug
  # Or use the native builder:
  # - out: src/db
  #   plugin: native
  #   options:
  #     import_extension: ".js"
```

### Plugin Options

| Option | Builder | Type | Required | Default | Description |
|--------|---------|------|----------|---------|-------------|
| `disable_template_literals` | Effect | boolean | No | `false` | Preserve original sqlc SQL using `sql.unsafe()` instead of transforming to template literals. See [Preserving Original SQL](#preserving-original-sql). |
| `import_extension` | Effect, native | string | No | Effect: `""`, native: `.js` | Explicit extension for generated relative imports. Allowed: `""`, `.js`, `.ts`. Use `.js` for Node ESM (`moduleResolution: nodenext`/`node16`). |
| `driver` | native | string | No | `pg` | Database driver target. Currently only `pg` is supported. |
| `validator` | native | string | No | `zod` | Runtime validator target. Currently only `zod` is supported. |
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
├── cmd/effect/           # Effect plugin source code
│   ├── main.go           # Entry point
│   └── internal/
│       ├── config/       # Plugin configuration
│       └── effect4/      # Effect v4 code generation
├── cmd/native/           # Native TypeScript plugin source code
│   ├── main.go           # Entry point
│   └── internal/
│       ├── config/       # Plugin configuration
│       └── native/       # Native TypeScript code generation
├── toolbelt/             # Shared sqlc mapping/generation helpers
│   ├── mapper/           # sqlc to IR mapping
│   ├── models/           # Public intermediate representation
│   └── logger/           # Structured logging
├── examples/             # Example projects
│   ├── effect-v4/        # Effect v4 example
│   └── native/           # Native TypeScript example
└── dist/                 # Built plugin artifacts
```

## License

See [LICENSE](LICENSE) file.
