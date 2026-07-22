# Effect v4 Plugin

The Effect v4 plugin generates idiomatic Effect v4 code using the `effect/unstable/sql` module.

## Requirements

- [Effect](https://effect.website) v4 (beta)
- TypeScript 5.5+

## What it generates

- **Type-safe parameter schemas** using Effect's Schema library
- **Type-safe result schemas** with proper null handling via `Option`
- **Repository services** with Effect's dependency injection via `Layer`
- **Automatic SQL type mapping** to TypeScript/Effect types

## Example

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

## Nested Results with `sqlc.embed`

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
// Generated automatically - Row schema decodes the flat database result
export const GetOrderWithCustomerRow = Schema.Struct({
  orders_id: Schema.Int,
  orders_customer_id: Schema.Int,
  orders_status: OrderStatusSchema,
  orders_total_cents: Schema.Int,
  orders_shipping_address: Schema.OptionFromNullOr(Schema.String),
  // ... more orders columns
  customers_id: Schema.Int,
  customers_email: Schema.String,
  customers_name: Schema.String,
  // ... more customers columns
})

export type GetOrderWithCustomerRow = typeof GetOrderWithCustomerRow.Type

// The public result schema describes the nested representation.
export const GetOrderWithCustomerResult = Schema.Struct({
  order: OrdersRow,
  customer: CustomersRow,
})

export type GetOrderWithCustomerResult = typeof GetOrderWithCustomerResult.Type

// sqlc.embed is a one-way projection from a decoded row to the public result.
export const mapGetOrderWithCustomerRowToResult = (
  row: GetOrderWithCustomerRow,
): GetOrderWithCustomerResult => ({
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
})

const findRow = SqlSchema.findOneOption({
  Request: GetOrderWithCustomerParams,
  Result: GetOrderWithCustomerRow,
  execute: (params) => sql`/* ... */`,
})

const getOrderWithCustomer = (params: GetOrderWithCustomerParams) =>
  findRow(params).pipe(
    Effect.map(Option.map(mapGetOrderWithCustomerRowToResult)),
  )

// Result type is nested:
// {
//   order: { id: number, status: "pending" | "shipped" | ..., shipping_address: Option<string>, ... }
//   customer: { id: number, email: string, name: string, ... }
// }
```

**Key features:**
- Table names are singularized for field names (`orders` -> `order`, `customers` -> `customer`)
- Columns are prefixed with table name to avoid conflicts (`orders_id`, `customers_id`)
- Enum types are preserved in embed schemas
- Nullable fields use `Schema.OptionFromNullOr` for consistent API with non-embed queries
- Flat rows are schema-decoded before a typed, one-way projection builds the nested result

## SQL Generation

By default, the plugin transforms sqlc's parameterized SQL into Effect's tagged template literal syntax:

```typescript
// Default output (template literals)
// GetCustomer
// SELECT * FROM customers WHERE id = $1 AND email = $2
execute: (params) => sql`SELECT * FROM customers WHERE id = ${params.id} AND email = ${params.email}`
```

The original SQL query is included as a comment above each query implementation for reference.

### Preserving Original SQL

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

## Repository Pattern

Each SQL file in your `queries/` directory becomes its own encapsulated repository. For example:

```
queries/
├── customers.sql    -> CustomersRepository.ts
├── orders.sql       -> OrdersRepository.ts
└── products.sql     -> ProductsRepository.ts
```

All queries defined in a SQL file are grouped into a single repository service. This keeps related database operations together and provides clean dependency injection through Effect's `Layer` system.

### Generated Output

For each repository, the builder generates:

- **Parameter schemas** - Type-safe input validation for each query
- **Result schemas** - Type-safe output parsing with proper null handling (`Option`)
- **Repository interface** - A typed interface defining all available operations
- **Service tag** - An Effect service tag for dependency injection
- **Live implementation** - The actual repository implementation using `SqlClient`
- **Layer export** - A ready-to-use `Layer` for providing the repository

### Usage

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

## Supported sqlc Commands

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

## Supported sqlc Macros

| Macro | Supported | Description |
|-------|-----------|-------------|
| `sqlc.arg('name')` | Yes | Explicit parameter naming |
| `@name` | Yes | PostgreSQL shorthand for `sqlc.arg(name)` |
| `sqlc.narg('name')` | Yes | Explicit nullable parameter naming |
| `sqlc.slice('name')` | Yes | Dynamic `IN` list expansion using Effect SQL's `sql.in` helper |
| `sqlc.embed(table)` | Yes | Embed table columns into nested structures |

`sqlc.slice` generates array request schemas and expands to Effect SQL's column-aware `sql.in` helper. Empty slices compile to a false predicate through Effect SQL.

## Configuration

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `disable_template_literals` | boolean | No | `false` | Preserve original sqlc SQL using `sql.unsafe()` instead of transforming to template literals. See [Preserving Original SQL](#preserving-original-sql). |
| `import_extension` | string | No | `""` | Explicit extension for generated relative imports. Allowed: `""`, `.js`, `.ts`. Use `.js` for Node ESM (`moduleResolution: nodenext`/`node16`). |
| `debug` | boolean | No | `false` | Enable debug mode to output intermediate representations and detailed logs during code generation. |
| `debug_dir` | string | No | `"debug"` | Directory where debug output files are written when debug mode is enabled. |
