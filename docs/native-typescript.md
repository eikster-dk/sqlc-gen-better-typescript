# Native TypeScript Plugin

The native plugin generates plain TypeScript functions with Zod validation and no Effect dependency. It targets PostgreSQL clients that implement a small `SqlClient` interface.

## Requirements

- TypeScript 5.0+
- [Zod](https://zod.dev) v3
- A PostgreSQL client implementing the `SqlClient` interface (e.g., `pg.Pool`, `pg.Client`)

## What it generates

- **Plain async query functions** grouped by SQL file
- **Zod request/result schemas** and inferred TypeScript types
- **A small database client interface** compatible with `pg.Pool` / `pg.Client`
- **Shared enum schemas** in `models.ts`
- **Nested `sqlc.embed` result transforms** for joined table structures

## Example

Write a SQL query:

```sql
-- name: GetCustomer :one
SELECT id, email, name, phone, created_at
FROM customers
WHERE id = $1;
```

Get a typed async function:

```typescript
const result = await getCustomer(pool, { id: 1 })

if (result.success) {
  console.log(result.data)
}
```

### Generated Client Interface

```typescript
export interface SqlClient {
  query(queryText: string, values?: unknown[]): Promise<SqlQueryResult>
}
```

### Return Type

Native query functions return `QueryResult<T>`:

```typescript
export type QueryResult<T> =
  | { success: true; data: T }
  | { success: false; error: ZodError; phase: "input" | "output" }
```

## File Structure

For each SQL file, the native builder generates three files:

```
queries/
├── customers.sql    -> customersRequests.ts, customersResponses.ts, customersQueries.ts
├── orders.sql       -> ordersRequests.ts, ordersResponses.ts, ordersQueries.ts
└── products.sql     -> productsRequests.ts, productsResponses.ts, productsQueries.ts
```

- `*Requests.ts`: Zod input schemas
- `*Responses.ts`: Zod row/result schemas
- `*Queries.ts`: async query functions

## Supported sqlc Commands

| Command | Supported | Native Return Type | Description |
|---------|-----------|--------------------|-------------|
| `:one` | Yes | `QueryResult<Result \| null>` | Returns at most one row |
| `:many` | Yes | `QueryResult<Result[]>` | Returns zero or more rows |
| `:exec` | Yes | `QueryResult<void>` | Executes without returning data |
| `:execrows` | Yes | `QueryResult<number>` | Returns the number of affected rows |
| `:execresult` | Yes | `QueryResult<SqlExecResult>` | Returns command tag and affected row count |
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
| `sqlc.slice('name')` | Yes | Dynamic `IN` list expansion; empty slices become `IN (NULL)` |
| `sqlc.embed(table)` | Yes | Embed table columns into nested result objects |

## Configuration

| Option | Type | Required | Default | Description |
|--------|------|----------|---------|-------------|
| `driver` | string | No | `pg` | Database driver target. Currently only `pg` is supported. |
| `validator` | string | No | `zod` | Runtime validator target. Currently only `zod` is supported. |
| `import_extension` | string | No | `.js` | Explicit extension for generated relative imports. Allowed: `""`, `.js`, `.ts`. |
| `debug` | boolean | No | `false` | Enable debug mode to output intermediate representations and detailed logs during code generation. |
| `debug_dir` | string | No | `"debug"` | Directory where debug output files are written when debug mode is enabled. |
