# sqlc-gen-effect

A [sqlc](https://sqlc.dev) WASM plugin that generates type-safe TypeScript code from your SQL queries.

## Requirements

- [sqlc](https://sqlc.dev) v1.25.0 or later
- PostgreSQL

## What is this?

**sqlc-gen-effect** contains two sqlc WASM plugins backed by a shared `toolbelt` package that normalizes sqlc plugin requests into a stable intermediate representation. Instead of writing boilerplate database access code, you write SQL and generate typed TypeScript code for either Effect v4 or plain async functions.

## Plugins

| Plugin | Description | Docs |
|--------|-------------|------|
| **Effect v4** | Repository services using `effect/unstable/sql` and Effect Schema | [docs/effect-v4.md](docs/effect-v4.md) |
| **Native TypeScript** | Plain async functions using Zod validation and a `SqlClient` interface | [docs/native-typescript.md](docs/native-typescript.md) |

### Planned

| Plugin | Description |
|--------|-------------|
| `effect-v3` | Effect v3 compatible code generation |

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

- **Parameters**: Nullable parameters use `Schema.OptionFromNullOr()`, keeping the field required while encoding `Option.none()` as SQL `NULL`
- **Results**: Nullable results use `Schema.OptionFromNullOr()`, transforming `null` to `Option.none()`

## Quick Start

1. Install sqlc: https://docs.sqlc.dev/en/latest/overview/install.html

2. Create your `sqlc.yaml` configuration:

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
      import_extension: ".js"
  # Or use the native builder:
  # - out: src/db
  #   plugin: native
  #   options:
  #     import_extension: ".js"
```

3. Write your SQL schema and queries

4. Run sqlc:
   ```bash
   sqlc generate
   ```

5. Use the generated code in your application

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
├── docs/                 # Plugin documentation
├── examples/             # Example projects
│   ├── effect-v4/        # Effect v4 example
│   └── native/           # Native TypeScript example
└── dist/                 # Built plugin artifacts
```

## License

See [LICENSE](LICENSE) file.
